package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/src-d/go-billy.v3/osfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"io/ioutil"

	"net/url"

	"regexp"

	"os/exec"

	"github.com/badgerodon/grpcsimulator/builder/builderpb"
	"github.com/cespare/xxhash"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// A Server builds go programs using gopherjs
type Server struct {
	workingDir string
	projectID  string
	bucket     string
}

// NewServer creates a new Server
func NewServer(workingDir, projectID, bucket string) *Server {
	return &Server{
		workingDir: workingDir,
		projectID:  projectID,
		bucket:     bucket,
	}
}

// Build builds the passed in repository
func (s *Server) Build(ctx context.Context, req *builderpb.BuildRequest) (*builderpb.BuildResponse, error) {
	importPathParts := strings.Split(req.ImportPath, "/")
	if len(importPathParts) < 3 {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid import path: %s", req.ImportPath)
	}
	if importPathParts[0] != "github.com" {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid import path: %s. only github.com or bitbucket.com is supported at this time", req.ImportPath)
	}

	provider, organization, repository := importPathParts[0], importPathParts[1], importPathParts[2]
	subfolder := strings.Join(importPathParts[3:], "/")

	log.Printf("building provider=%s organization=%s repository=%s subfolder=%s\n",
		provider, organization, repository, subfolder)

	dir := getSafeName(req.ImportPath)

	tmp, err := ioutil.TempDir(s.workingDir, dir)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to create temporary directory: %v", err)
	}

	checkoutPath := filepath.Join(tmp, "src", provider, organization, repository)
	err = os.MkdirAll(checkoutPath, 0755)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to create src folder: %v", err)
	}

	_, err = git.Clone(memory.NewStorage(), osfs.New(checkoutPath), &git.CloneOptions{
		URL:           "https://" + url.PathEscape(provider) + "/" + url.PathEscape(organization) + "/" + url.PathEscape(repository) + ".git",
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

	cmd := exec.Command("gopherjs", "build", "-o", filepath.Join(binPath, "bin.js"), req.ImportPath)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", tmp))
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to build code: %v\n%v", err, string(bs))
	}

	return &builderpb.BuildResponse{
		Location: filepath.Join(binPath, "bin.js"),
	}, nil
}

type buildContext struct {
	sync.Mutex
	dir        string
	importPath string
	branch     string
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
