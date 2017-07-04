package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-billy.v3/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"io/ioutil"

	"net/url"

	"regexp"

	"os/exec"

	"net/http"

	"encoding/json"

	"cloud.google.com/go/storage"
	"github.com/badgerodon/grpcsimulator/builder/builderpb"
	"github.com/cespare/xxhash"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	projectID  string
	bucket     string
	folder     string

	client *storage.Client
}

// NewServer creates a new Server
func NewServer(workingDir, projectID, bucket, folder string) (*Server, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &Server{
		workingDir: workingDir,
		projectID:  projectID,
		bucket:     bucket,
		folder:     folder,

		client: client,
	}, nil
}

// Build builds the passed in repository
func (s *Server) Build(ctx context.Context, req *builderpb.BuildRequest) (*builderpb.BuildResponse, error) {
	importPathParts := strings.Split(req.ImportPath, "/")
	if len(importPathParts) < 3 {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid import path: %s", req.ImportPath)
	}

	ref := vcsReference{
		provider:     importPathParts[0],
		organization: importPathParts[1],
		repository:   importPathParts[2],
		branch:       req.Branch,
	}
	dir := getSafeName(req.ImportPath)

	head, err := s.getHeadCommit(ctx, ref)
	if err != nil {
		return nil, err
	}

	fileName := head + ".js"
	remotePath := path.Join(s.folder, dir, fileName)
	webPath := "https://storage.googleapis.com/" + s.bucket + "/" + s.folder + "/" + dir + "/" + fileName
	exists, err := s.objectExists(ctx, remotePath)
	if err != nil {
		return nil, err
	}
	if exists {
		return &builderpb.BuildResponse{
			Location: webPath,
		}, nil
	}

	tmp, err := ioutil.TempDir(s.workingDir, dir)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to create temporary directory: %v", err)
	}

	checkoutPath := filepath.Join(tmp, "src", ref.provider, ref.organization, ref.repository)
	err = os.MkdirAll(checkoutPath, 0755)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to create src folder: %v", err)
	}

	_, err = git.Clone(memory.NewStorage(), osfs.New(checkoutPath), &git.CloneOptions{
		URL:           ref.gitURL(),
		ReferenceName: plumbing.ReferenceName("refs/heads/" + req.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to clone git repository: %v", err)
	}

	binPath := filepath.Join(tmp, "bin")
	err = os.MkdirAll(binPath, 0755)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to create bin folder: %v", err)
	}

	cmd := exec.Command("gopherjs", "build", "-o", filepath.Join(binPath, fileName), req.ImportPath)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", tmp))
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to build code: %v\n%v", err, string(bs))
	}

	for remotePath, localPath := range map[string]string{
		remotePath:          filepath.Join(binPath, fileName),
		remotePath + ".map": filepath.Join(binPath, fileName+".map"),
	} {
		err = s.uploadFile(ctx, remotePath, localPath)
		if err != nil {
			return nil, grpc.Errorf(codes.Unknown, "failed to upload file: %v", err)
		}
	}

	return &builderpb.BuildResponse{
		Location: webPath,
	}, nil
}

func (s *Server) getHeadCommit(ctx context.Context, ref vcsReference) (string, error) {
	switch ref.organization {
	case "github.com":
		u := "https://api.github.com/repos/" + url.PathEscape(ref.organization) +
			"/" + url.PathEscape(ref.repository) +
			"/git/refs/" + url.PathEscape(ref.branch)
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

		var obj struct {
			Ref    string `json:"ref"`
			URL    string `json:"url"`
			Object struct {
				Type string `json:"type"`
				SHA  string `json:"sha"`
				URL  string `json:"url"`
			} `json:"object"`
		}
		err = json.NewDecoder(res.Body).Decode(&obj)
		if err != nil {
			return "", err
		}

		return obj.Object.SHA, nil
	default:
		return "", errors.New("unsupported provider")
	}
}

func (s *Server) objectExists(ctx context.Context, remotePath string) (exists bool, err error) {
	object := s.client.Bucket(s.bucket).Object(remotePath)
	_, err = object.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	} else if err != nil {
		return false, grpc.Errorf(codes.Unknown, "failed to get object attributes: %v remote_path=%s",
			err, remotePath)
	}

	return true, nil
}

func (s *Server) uploadFile(ctx context.Context, remotePath, localPath string) error {
	src, err := os.Open(localPath)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to open source file: %v remote_path=%s local_path=%s",
			err, remotePath, localPath)
	}
	defer src.Close()

	object := s.client.Bucket(s.bucket).Object(remotePath)
	objectWriter := object.NewWriter(ctx)
	defer objectWriter.Close()

	_, err = io.Copy(objectWriter, src)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to upload file: %v remote_path=%s local_path=%s",
			err, remotePath, localPath)
	}

	err = objectWriter.Close()
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to upload file: %v remote_path=%s local_path=%s",
			err, remotePath, localPath)
	}

	return nil
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

func getSafeName(original string) string {
	h := xxhash.New()
	io.WriteString(h, original)
	suffix := hex.EncodeToString(h.Sum(nil))
	prefix := unsafeCharacters.ReplaceAllLiteralString(original, "-")
	if len(prefix) > 500 {
		prefix = prefix[:500]
	}
	return prefix + "--" + suffix
}
