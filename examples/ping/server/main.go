package main

import (
	"log"
	"net"

	"github.com/badgerodon/grpcsimulator/examples/ping/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	_ "github.com/badgerodon/grpcsimulator/kernel"
)

func main() {
	li, err := net.Listen("tcp", "127.0.0.1:7000")
	if err != nil {
		panic(err)
	}
	defer li.Close()

	go func() {
		cc, err := grpc.Dial("127.0.0.1:7000", grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		c := pb.NewPingServiceClient(cc)
		for i := 0; i < 100; i++ {
			req := &pb.PingRequest{
				Message: "Hello World",
			}

			res, err := c.Ping(context.Background(), req)
			log.Printf("request=%v result=%v error=%v\n", req, res, err)
		}

	}()

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
