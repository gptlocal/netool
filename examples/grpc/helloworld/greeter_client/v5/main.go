package main

import (
	"context"
	pb "github.com/gptlocal/netool/examples/grpc/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"time"
)

func main() {
	ka := keepalive.ClientParameters{
		Time:                10 * time.Minute,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(ka))
	if err != nil {
		log.Printf("Did not connect: %v", err)
		return
	}
	defer conn.Close()

	in := &healthpb.HealthCheckRequest{
		Service: pb.Greeter_SayHello_FullMethodName,
	}

	c := healthpb.NewHealthClient(conn)
	stream, err := c.Watch(context.Background(), in)
	if err != nil {
		log.Fatalf("Failed to watch health status: %v", err)
	}

	time.Sleep(2 * time.Minute)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		log.Printf("Health status: %v", resp.GetStatus())
	}
}
