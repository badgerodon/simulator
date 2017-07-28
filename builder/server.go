package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	dir := getSafeName(req.ImportPath, head)
	dst := filepath.Join(s.workingDir, dir, filepath.Base(req.ImportPath)+".js")

	s.mu.Lock()
	s.lastAccess[dst] = time.Now()
	s.mu.Unlock()

	_, err, _ = s.singleflight.Do(dst, func() (interface{}, error) {
		var err error
		//if _, err = os.Stat(dst); err != nil {
		err = s.buildVCS(dst, ref, req, head)
		//}
		return nil, err
	})
	if err != nil {
		return nil, err
	}

	return &builderpb.BuildResponse{
		Location: dst,
	}, nil
}

func (s *Server) buildVCS(dst string, ref vcsReference, req *builderpb.BuildRequest, head string) error {
	tmp, err := ioutil.TempDir(s.workingDir, getSafeName(req.ImportPath, head))
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmp)

	checkoutPath := filepath.Join(tmp, "src", ref.provider, ref.organization, ref.repository)
	err = os.MkdirAll(checkoutPath, 0755)
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

	binPath := filepath.Join(tmp, "bin")
	err = os.MkdirAll(binPath, 0755)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to create bin folder: %v", err)
	}

	cmd := exec.Command("gopherjs", "build", "-o", filepath.Join(binPath, filepath.Base(dst)), req.ImportPath)
	gopaths := []string{tmp}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			gopaths = append([]string{e[7:]}, gopaths...)
		} else {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", strings.Join(gopaths, ":")))
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to build code: %v\n%v", err, string(bs))
	}

	// post process the js file to inline the map
	f, err := os.Open(filepath.Join(binPath, filepath.Base(dst)))
	if err != nil {
		return err
	}
	defer f.Close()
	os.MkdirAll(filepath.Dir(dst), 0755)

	err = os.Rename(filepath.Join(binPath, filepath.Base(dst)), dst)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to move file to working directory: %v", err)
	}

	mapPath := filepath.Join(binPath, filepath.Base(dst)+".map")
	err = s.injectMapSources(mapPath+".out", mapPath)
	if err != nil {
		return err
	}

	err = os.Rename(mapPath+".out", dst+".map")
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to move file to working directory: %v", err)
	}

	return nil
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
		Version        int       `json:"version"`
		File           string    `json:"file"`
		SourceRoot     string    `json:"sourceRoot"`
		Sources        []string  `json:"sources"`
		SourcesContent []*string `json:"sourcesContent"`
		Names          []string  `json:"names"`
		Mappings       string    `json:"mappings"`
	}
	err = json.NewDecoder(sf).Decode(&sourceMap)
	if err != nil {
		return err
	}

	sourceMap.SourcesContent = make([]*string, 0, len(sourceMap.Sources))
	for _, sourcePath := range sourceMap.Sources {
		candidates := []string{
			filepath.Join(os.Getenv("GOROOT"), "src", sourcePath),
			filepath.Join(os.Getenv("GOPATH"), "src", sourcePath),
		}
		var content *string
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				bs, _ := ioutil.ReadFile(candidate)
				content = new(string)
				*content = string(bs)
			}
		}

		sourceMap.SourcesContent = append(sourceMap.SourcesContent, content)
	}

	log.Println("MAP", sourceMap)

	return json.NewEncoder(df).Encode(&sourceMap)
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
