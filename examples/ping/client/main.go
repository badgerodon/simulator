package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"github.com/badgerodon/grpcsimulator/examples/ping/pb"

	_ "github.com/badgerodon/grpcsimulator/kernel"
)

func main() {
	log.SetFlags(0)

	conn, err := grpc.Dial("127.0.0.1:7000",
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
