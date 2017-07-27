package router

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/badgerodon/simulator/mesh/meshpb"
	"github.com/hashicorp/yamux"
)

type (
	// A Mesh coordinates communication between clients and servers
	Mesh struct {
		cfg *Config

		mu       sync.Mutex
		sessions map[string]*yamux.Session
	}
)

var (
	ackOK = meshpb.Message{
		Type: meshpb.MessageType_ACK,
		Data: []byte{1},
	}
	ackFail = meshpb.Message{
		Type: meshpb.MessageType_ACK,
		Data: []byte{0},
	}
)

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

// Handle connects a connection to the mesh network
func (m *Mesh) Handle(conn net.Conn) {
	defer conn.Close()

	var req meshpb.Message
	err := meshpb.Read(conn, &req)
	if err != nil {
		log.Println("invalid handshake", err)
		return
	}

	switch req.Type {
	case meshpb.MessageType_DIAL:
		endpoint := string(req.Data)

		m.mu.Lock()
		session, ok := m.sessions[endpoint]
		m.mu.Unlock()

		if !ok {
			meshpb.Write(conn, &ackFail)
			return
		}

		stream, err := session.Open()
		if err != nil {
			meshpb.Write(conn, &ackFail)
			log.Println("failed to open stream", err)
			return
		}

		err = meshpb.Write(conn, &ackOK)
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
	case meshpb.MessageType_LISTEN:
		endpoint := string(req.Data)

		m.mu.Lock()
		_, ok := m.sessions[endpoint]
		m.mu.Unlock()

		if ok {
			meshpb.Write(conn, &ackFail)
			return
		}

		session, err := yamux.Server(conn, nil)
		if err != nil {
			m.mu.Unlock()
			meshpb.Write(conn, &ackFail)
			log.Println("failed to create session", err)
			return
		}

		err = meshpb.Write(conn, &ackOK)
		if err != nil {
			return
		}

		m.mu.Lock()
		cur, ok := m.sessions[endpoint]
		if ok {
			// re-use the existing session and close this one
			session.Close()
		} else {
			// add this session to the map
			m.sessions[endpoint] = session
			cur = session
		}
		m.mu.Unlock()

		// if the connection is closed, remove it from the map
		for range time.Tick(time.Millisecond) {
			if cur.IsClosed() {
				m.mu.Lock()
				session, ok := m.sessions[endpoint]
				if ok && session == cur {
					delete(m.sessions, endpoint)
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
