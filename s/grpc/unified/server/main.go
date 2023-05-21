package main

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
)

const (
	port = 28800
)

func main() {
	gs := grpc.NewServer(grpc.UnknownServiceHandler(UnifiedHandler))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicf("fail to listen on %d, err:%v\n", port, err)
	}

	log.Printf("Server listening on %s\n", lis.Addr())
	if e := gs.Serve(lis); e != nil {
		log.Panicf("fail to serve with err: %v\n", e)
	}
}

func UnifiedHandler(srv interface{}, stream grpc.ServerStream) error {
	log.Printf("srv: %v, stream:%v, context: %v\n", srv, stream, stream.Context())

	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		log.Printf("Can't find Method in ServerStream")
		return errors.New("method not found")
	}

	log.Printf("FullMethodName: %s\n", fullMethodName)

	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Printf("Can't extract metadata")
		return nil
	}
	log.Printf("Metadata: %v\n", md)

	return nil
}
