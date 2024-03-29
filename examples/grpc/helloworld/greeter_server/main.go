package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/gptlocal/netool/examples/grpc/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"time"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, v := range md {
			log.Printf("Received header metadata : %v : %v", k, v)
		}
	}
	err := grpc.SetTrailer(ctx, metadata.Pairs("tracing-enabled", "1"))
	if err != nil {
		log.Printf("SetTrailer error: %v", err)
	}
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	kp := keepalive.ServerParameters{
		MaxConnectionIdle: 24 * time.Hour,
		Time:              2 * time.Hour,
		Timeout:           20 * time.Second,
	}
	s := grpc.NewServer(grpc.KeepaliveParams(kp))
	//s := grpc.NewServer()

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	pb.RegisterGreeterServer(s, &server{})

	reflection.Register(s)
	healthServer.SetServingStatus(pb.Greeter_SayHello_FullMethodName, grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
