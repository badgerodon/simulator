package mesh

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/badgerodon/grpcsimulator/mesh/meshpb"
	"github.com/hashicorp/yamux"
)

type (
	// A Mesh coordinates communication between clients and servers
	Mesh struct {
		cfg *Config

		mu       sync.Mutex
		sessions map[string]*yamux.Session
	}
	// A Config configures the mesh
	Config struct {
		Address string
	}
)

// DefaultConfig is the default Config
var DefaultConfig = &Config{
	Address: "127.0.0.1:7788",
}

// New creates a new mesh
func New(options ...func(*Config)) *Mesh {
	cfg := new(Config)
	*cfg = *DefaultConfig

	for _, f := range options {
		f(cfg)
	}

	return &Mesh{
		cfg:      cfg,
		sessions: make(map[string]*yamux.Session),
	}
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

func (m *Mesh) handle(conn net.Conn) {
	defer conn.Close()

	var req meshpb.HandshakeRequest
	err := meshpb.Read(conn, &req)
	if err != nil {
		log.Println("invalid handshake", err)
		return
	}

	switch req.Type {
	case meshpb.HandshakeType_DIAL:
		m.mu.Lock()
		session, ok := m.sessions[req.Endpoint]
		m.mu.Unlock()

		if !ok {
			meshpb.Write(conn, &meshpb.HandshakeResult{
				Status: meshpb.HandshakeStatus_FAILED,
			})
			return
		}

		stream, err := session.Open()
		if err != nil {
			meshpb.Write(conn, &meshpb.HandshakeResult{
				Status: meshpb.HandshakeStatus_FAILED,
			})
			log.Println("failed to open stream", err)
			return
		}

		err = meshpb.Write(conn, &meshpb.HandshakeResult{
			Status: meshpb.HandshakeStatus_CONNECTED,
		})
		if err != nil {
			return
		}

		// if either sending or receive fails, close the connection
		errs := make(chan error, 2)
		go func() {
			_, err := io.Copy(conn, stream)
			errs <- err
		}()
		go func() {
			_, err := io.Copy(stream, conn)
			errs <- err
		}()
		<-errs
	case meshpb.HandshakeType_LISTEN:
		m.mu.Lock()
		_, ok := m.sessions[req.Endpoint]
		m.mu.Unlock()

		if ok {
			meshpb.Write(conn, &meshpb.HandshakeResult{
				Status: meshpb.HandshakeStatus_FAILED,
			})
			return
		}

		session, err := yamux.Server(conn, nil)
		if err != nil {
			m.mu.Unlock()
			meshpb.Write(conn, &meshpb.HandshakeResult{
				Status: meshpb.HandshakeStatus_FAILED,
			})
			log.Println("failed to create session", err)
			return
		}

		err = meshpb.Write(conn, &meshpb.HandshakeResult{
			Status: meshpb.HandshakeStatus_CONNECTED,
		})
		if err != nil {
			return
		}

		m.mu.Lock()
		cur, ok := m.sessions[req.Endpoint]
		if ok {
			// re-use the existing session and close this one
			session.Close()
		} else {
			// add this session to the map
			m.sessions[req.Endpoint] = session
			cur = session
		}
		m.mu.Unlock()

		// if the connection is closed, remove it from the map
		for range time.Tick(time.Millisecond) {
			if cur.IsClosed() {
				m.mu.Lock()
				session, ok := m.sessions[req.Endpoint]
				if ok && session == cur {
					delete(m.sessions, req.Endpoint)
				}
				m.mu.Unlock()
				break
			}
		}

	default:
		log.Println("unknown handshake type", req.Type)
		return
	}
}
