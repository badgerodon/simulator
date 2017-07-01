// +build !js

package mesh

import (
	"net"
	"time"
)

type (
	// A Config configures the mesh
	Config struct {
		Address string
	}
)

// DefaultConfig is the default Config
var DefaultConfig = &Config{
	Address: "127.0.0.1:7788",
}

// Serve serves request for the mesh
func (m *Mesh) Serve() error {
	li, err := net.Listen("tcp", m.cfg.Address)
	if err != nil {
		return err
	}

	for {
		conn, err := li.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(1)
				continue
			}
			return err
		}
		go m.handle(conn)
	}
}
