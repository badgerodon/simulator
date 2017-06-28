package ping

import (
	"github.com/badgerodon/grpcsimulator/examples/ping/pb"
	"golang.org/x/net/context"
)

type Server struct {
}

func New() pb.PingServiceServer {
	return &Server{}
}

func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: req.GetMessage(),
	}, nil
}
