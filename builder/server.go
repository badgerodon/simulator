package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
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

	"cloud.google.com/go/storage"
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
	folder     string
}

// NewServer creates a new Server
func NewServer(workingDir, projectID, bucket, folder string) *Server {
	return &Server{
		workingDir: workingDir,
		projectID:  projectID,
		bucket:     bucket,
		folder:     folder,
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

	repo, err := git.Clone(memory.NewStorage(), osfs.New(checkoutPath), &git.CloneOptions{
		URL:           "https://" + url.PathEscape(provider) + "/" + url.PathEscape(organization) + "/" + url.PathEscape(repository) + ".git",
		ReferenceName: plumbing.ReferenceName("refs/heads/" + req.Branch),
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "failed to clone git repository: %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	fileName := head.Hash().String() + ".js"
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

func (s *Server) objectExists(ctx context.Context, remotePath string) (exists bool, err error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, grpc.Errorf(codes.Unknown, "failed to create storage client: %v remote_path=%s",
			err, remotePath)
	}
	defer client.Close()

	object := client.Bucket(s.bucket).Object(remotePath)
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

	client, err := storage.NewClient(ctx)
	if err != nil {
		return grpc.Errorf(codes.Unknown, "failed to create storage client: %v remote_path=%s local_path=%s",
			err, remotePath, localPath)
	}
	defer client.Close()

	object := client.Bucket(s.bucket).Object(remotePath)
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
