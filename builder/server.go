package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/badgerodon/simulator/builder/builderpb"
	"golang.org/x/net/context"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/src-d/go-billy.v3/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type vcsReference struct {
	provider     string
	organization string
	repository   string
	branch       string
}

func (ref vcsReference) gitURL() string {
	return "https://" + url.PathEscape(ref.provider) +
		"/" + url.PathEscape(ref.organization) +
		"/" + url.PathEscape(ref.repository) + ".git"
}

// A Server builds go programs using gopherjs
type Server struct {
	workingDir string

	singleflight singleflight.Group
	mu           sync.Mutex
	lastAccess   map[string]time.Time
}

// NewServer creates a new Server
func NewServer(workingDir string) (*Server, error) {
	return &Server{
		workingDir: workingDir,

		lastAccess: make(map[string]time.Time),
	}, nil
}

// Build builds the passed in repository
func (s *Server) Build(ctx context.Context, req *builderpb.BuildRequest) (*builderpb.BuildResponse, error) {
	importPathParts := strings.Split(req.ImportPath, "/")
	if len(importPathParts) < 3 {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid import path: %s",
			req.ImportPath)
	}

	ref := vcsReference{
		provider:     importPathParts[0],
		organization: importPathParts[1],
		repository:   importPathParts[2],
		branch:       req.Branch,
	}

	head, err := s.getHeadCommit(ctx, ref)
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(s.workingDir, getSafeName(req.ImportPath, head))

	s.mu.Lock()
	s.lastAccess[dir] = time.Now()
	s.mu.Unlock()

	_, err, _ = s.singleflight.Do(dir, func() (interface{}, error) {
		var err error
		//if _, err = os.Stat(dir); err != nil {
		err = s.buildVCS(ref, req, dir, head)
		//}
		return nil, err
	})
	if err != nil {
		return nil, err
	}

	return &builderpb.BuildResponse{
		Location: filepath.Join(dir, filepath.Base(req.ImportPath)+".js"),
	}, nil
}

func (s *Server) buildVCS(ref vcsReference, req *builderpb.BuildRequest, dir, head string) error {
	name := filepath.Base(req.ImportPath)

	checkoutPath := filepath.Join(dir, "src", ref.provider, ref.organization, ref.repository)
	err := os.MkdirAll(checkoutPath, 0755)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to create src folder: %v", err)
	}

	_, err = git.Clone(memory.NewStorage(), osfs.New(checkoutPath), &git.CloneOptions{
		URL:           ref.gitURL(),
		ReferenceName: plumbing.ReferenceName("refs/heads/" + req.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to clone git repository: %v", err)
	}

	fmt.Println(filepath.Join(dir, "src", req.ImportPath))
	err = writeFiles(filepath.Join(dir, "src", req.ImportPath))
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to write template files: %v", err)
	}

	// download dependencies if there is no vendor directory
	if _, e := os.Stat(filepath.Join(dir, "vendor")); e != nil {
		cmd := exec.Command("go", "get", "-d", req.ImportPath)
		cmd.Env = s.getEnv(dir)
		bs, err := cmd.CombinedOutput()
		if err != nil {
			return grpc.Errorf(codes.Unknown, "failed to get dependencies: %v\n%v", err, string(bs))
		}
	}

	cmd := exec.Command("gopherjs", "build", "-o", filepath.Join(dir, name+".js"), req.ImportPath)
	cmd.Env = s.getEnv(dir)
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to build code: %v\n%v", err, string(bs))
	}

	mapPath := filepath.Join(dir, name+".js.map")
	err = s.injectMapSources(mapPath+".out", mapPath)
	if err != nil {
		return err
	}
	err = os.Rename(mapPath+".out", mapPath)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to move file to working directory: %v", err)
	}

	return nil
}

func (s *Server) getEnv(dir string) []string {
	var env []string
	gopaths := []string{dir}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			gopaths = append(gopaths, e[7:])
		} else {
			env = append(env, e)
		}
	}
	return append(env, fmt.Sprintf("GOPATH=%s", strings.Join(gopaths, ":")))
}

func (s *Server) injectMapSources(dst, src string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()

	var sourceMap struct {
		Version    int      `json:"version"`
		File       string   `json:"file"`
		SourceRoot string   `json:"sourceRoot"`
		Sources    []string `json:"sources"`
		Names      []string `json:"names"`
		Mappings   string   `json:"mappings"`
	}
	err = json.NewDecoder(sf).Decode(&sourceMap)
	if err != nil {
		return err
	}

	for i, sourcePath := range sourceMap.Sources {
		candidates := []string{
			filepath.Join(os.Getenv("GOROOT"), "src", sourcePath),
			filepath.Join(os.Getenv("GOPATH"), "src", sourcePath),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				s.copyFile(
					filepath.Join(filepath.Dir(src), "src", sourcePath),
					candidate,
				)
			}
		}
		if strings.HasPrefix(sourcePath, "/") {
			sourceMap.Sources[i] = sourcePath[1:]
		}
		sourceMap.Sources[i] = "src/" + sourceMap.Sources[i]
	}

	return json.NewEncoder(df).Encode(&sourceMap)
}

func (s *Server) copyFile(dst, src string) error {
	os.MkdirAll(filepath.Dir(dst), 0755)

	dstf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstf.Close()

	srcf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcf.Close()

	_, err = io.Copy(dstf, srcf)
	return err
}

func (s *Server) getHeadCommit(ctx context.Context, ref vcsReference) (string, error) {
	switch ref.provider {
	case "github.com":
		u := "https://api.github.com/repos" +
			"/" + url.PathEscape(ref.organization) +
			"/" + url.PathEscape(ref.repository) +
			"/branches" +
			"/" + url.PathEscape(ref.branch)
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req = req.WithContext(ctx)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		var tmp bytes.Buffer

		var obj struct {
			Commit struct {
				SHA string `json:"sha"`
			} `json:"commit"`
		}
		err = json.NewDecoder(io.TeeReader(res.Body, &tmp)).Decode(&obj)
		if err != nil {
			return "", grpc.Errorf(codes.Unknown, "failed to find commit: %v\n%s", err, tmp.String())
		}

		return obj.Commit.SHA, nil
	default:
		return "", errors.New("unsupported provider")
	}
}

func getHash(args ...string) string {
	h := sha256.New()
	for _, arg := range args {
		io.WriteString(h, arg)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

var unsafeCharacters = regexp.MustCompile(`[^a-zA-Z0-9]`)

func getSafeName(original, head string) string {
	prefix := unsafeCharacters.ReplaceAllLiteralString(original, "-")
	if len(prefix) > 500 {
		prefix = prefix[:500]
	}
	return prefix + "--" + head
}
