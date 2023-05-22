package main

import (
	"context"
	pb "github.com/gptlocal/netool/examples/grpc/helloworld"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	WatchHealth(context.Background(), "localhost:50051", pb.Greeter_SayHello_FullMethodName)
}

func WatchHealth(ctx context.Context, address string, service string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			watchServiceHealth(ctx, address, service)
			// 在尝试重新连接之前稍作延迟，防止过于频繁的连接尝试。
			time.Sleep(time.Second)
		}
	}
}

func watchServiceHealth(ctx context.Context, address string, service string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("Did not connect: %v", err)
		return
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)

	in := &grpc_health_v1.HealthCheckRequest{
		Service: service,
	}

	stream, err := client.Watch(ctx, in)
	if err != nil {
		log.Printf("Failed to watch health status: %v", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Failed to receive a health status: %v", err)
			break
		}
		log.Printf("Health status: %v", resp.GetStatus())
	}
}
