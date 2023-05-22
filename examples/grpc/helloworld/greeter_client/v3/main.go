package main

import (
	"context"
	"flag"
	"google.golang.org/grpc/metadata"
	"log"
	"time"

	pb "github.com/gptlocal/netool/examples/grpc/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 发送header metadata
	md := metadata.Pairs("trace-id", "12345678901234567890123456789012:1234567890123456:9876543210123456")
	ctx = metadata.NewOutgoingContext(ctx, md)

	var tailer metadata.MD
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name}, grpc.Trailer(&tailer))
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Trailer: %v", tailer)
	log.Printf("Greeting: %s", r.GetMessage())
}
