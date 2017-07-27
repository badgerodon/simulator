package router

// import (
// 	"fmt"
// 	"net"
// 	"os"
// 	"time"

// 	"github.com/badgerodon/simulator/mesh/meshpb"
// 	"github.com/hashicorp/yamux"
// )

// func connect(timeout time.Duration) (net.Conn, error) {
// 	addr := os.Getenv("MESH_SERVER")
// 	if addr == "" {
// 		addr = "127.0.0.1:7788"
// 	}
// 	conn, err := net.DialTimeout("tcp", addr, timeout)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

// // Dial connects to the given endpoint
// func Dial(endpoint string, timeout time.Duration) (net.Conn, error) {
// 	conn, err := connect(timeout)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req := meshpb.HandshakeRequest{
// 		Type:     meshpb.HandshakeType_DIAL,
// 		Endpoint: endpoint,
// 	}
// 	err = meshpb.Write(conn, &req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var res meshpb.HandshakeResult
// 	err = meshpb.Read(conn, &res)
// 	if err != nil {
// 		return nil, err
// 	}
// 	switch res.Status {
// 	case meshpb.HandshakeStatus_FAILED:
// 		return nil, fmt.Errorf("failed to connect to: %s", endpoint)
// 	}

// 	return conn, nil
// }

// // Listen starts a new listener in the service mesh
// func Listen(endpoint string) (net.Listener, error) {
// 	conn, err := connect(0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req := meshpb.HandshakeRequest{
// 		Type:     meshpb.HandshakeType_LISTEN,
// 		Endpoint: endpoint,
// 	}
// 	err = meshpb.Write(conn, &req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var res meshpb.HandshakeResult
// 	err = meshpb.Read(conn, &res)
// 	if err != nil {
// 		return nil, err
// 	}
// 	switch res.Status {
// 	case meshpb.HandshakeStatus_FAILED:
// 		return nil, fmt.Errorf("failed to listen on: %s", endpoint)
// 	}

// 	server, err := yamux.Server(conn, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return server, nil
// }
