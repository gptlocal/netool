package main

import (
	"context"
	"flag"
	pb "github.com/gptlocal/netool/examples/grpc/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"io"
	"log"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := healthpb.NewHealthClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Check(ctx, &healthpb.HealthCheckRequest{Service: pb.Greeter_SayHello_FullMethodName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Status)

	in := &healthpb.HealthCheckRequest{
		Service: pb.Greeter_SayHello_FullMethodName,
	}

	stream, err := c.Watch(context.Background(), in)
	if err != nil {
		log.Fatalf("Failed to watch health status: %v", err)
	}

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
