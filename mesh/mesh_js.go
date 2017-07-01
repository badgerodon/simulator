// +build js

package mesh

import "github.com/gopherjs/gopherjs/js"

type (
	// A Config configures the mesh
	Config struct {
		*js.Object
	}
)

// DefaultConfig is the default Config
var DefaultConfig = &Config{
	Object: js.Global,
}

// Serve serves request for the mesh
func (m *Mesh) Serve() error {
	m.Object.Set("onmessage", func(evt *js.Object) {
		log(evt)
	})
}

func log(args ...interface{}) {
	js.Global.Get("console").Call("log", args...)
}
