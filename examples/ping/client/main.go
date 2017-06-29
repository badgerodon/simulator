package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/badgerodon/grpcsimulator/examples/ping/pb"
	"github.com/badgerodon/grpcsimulator/mesh"
)

func main() {
	log.SetFlags(0)

	conn, err := grpc.Dial("test-1",
		grpc.WithDialer(mesh.Dial),
		grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client := pb.NewPingServiceClient(conn)
	res, err := client.Ping(context.Background(), &pb.PingRequest{
		Message: "Hello World",
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(res)
}
