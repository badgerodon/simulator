package runnerjs

import (
	"fmt"
	"net/url"

	"github.com/badgerodon/simulator/mesh"
	"github.com/gopherjs/gopherjs/js"
)

// A WorkerConfig describes the program to run and the listener to start for a worker
type WorkerConfig struct {
	Name       string
	ImportPath string
	Branch     string
}

type Worker interface {
	Close() error
}

// A Runner runs a program and connects it to the mesh network
type Runner struct {
	mesh *mesh.Mesh
}

// NewRunner creates a new Runner
func NewRunner(mesh *mesh.Mesh) *Runner {
	return &Runner{
		mesh: mesh,
	}
}

func (r *Runner) StartWorker() Worker {

}

// Run starts a new worker
func (r *Runner) Run(endpoint, importPath, branch string) error {
	worker := js.Global.Get("Worker").New(fmt.Sprintf("/build?endpoint=%s&import_path=%s&branch=%s",
		url.QueryEscape(endpoint),
		url.QueryEscape(importPath),
		url.QueryEscape(branch),
	))

	return nil
}
