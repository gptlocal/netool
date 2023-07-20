package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"log"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("Did not connect: %v", err)
		return
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)

	in := &grpc_health_v1.HealthCheckRequest{
		Service: "/helloworld.Greeter/SayHello",
	}

	var tailer metadata.MD
	log.Printf("Trailer: %v", tailer)
	stream, err := client.Watch(context.Background(), in, grpc.Trailer(&tailer))
	if err != nil {
		log.Printf("Watch error: %v", err)
		return
	}

	requests := []*grpc_health_v1.HealthCheckRequest{
		{
			Service: "/helloworld.Greeter/SayHello",
		},
		{
			Service: "/helloworld.Greeter/SayHello",
		},
		{
			Service: "/helloworld.Greeter/SayHello",
		},
		{
			Service: "/helloworld.Greeter/SayHello",
		},
		{
			Service: "/helloworld.Greeter/SayHello",
		},
	}

	// todo
	for _, req := range requests {
		if err := stream.SendMsg(req); err != nil {
			log.Printf("SendMsg error: %v", err)
			continue
		}
	}
	log.Printf("Trailer: %v", tailer)

	//if err := stream.CloseSend(); err != nil {
	//	log.Printf("CloseSend error: %v", err)
	//}

}
