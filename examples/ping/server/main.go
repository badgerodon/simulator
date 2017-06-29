package main

import (
	"log"

	"github.com/badgerodon/grpcsimulator/examples/ping/pb"
	"github.com/badgerodon/grpcsimulator/mesh"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	li, err := mesh.Listen("test-1")
	if err != nil {
		log.Fatalln("failed to start mesh listener", err)
	}
	defer li.Close()

	server := grpc.NewServer()
	pb.RegisterPingServiceServer(server, New())
	err = server.Serve(li)
	if err != nil {
		log.Fatalln("failed to serve requests on listener", err)
	}
}

// Server implements ping
type Server struct {
}

// New creates a new Ping server
func New() pb.PingServiceServer {
	return &Server{}
}

// Ping responds with the message sent to it
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: req.GetMessage(),
	}, nil
}
