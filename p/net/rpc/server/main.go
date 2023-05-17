package main

import (
	"github.com/gptlocal/netool/p/net/rpc"
	"github.com/gptlocal/netool/p/net/rpc/service"
	"log"
	"net"
	"net/http"
)

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting tcp server: %v", err)
	}

	newServer := rpc.NewServer()
	newServer.Register(new(service.Arith))

	handler := http.NewServeMux()
	handler.Handle(rpc.DefaultRPCPath, newServer)
	handler.Handle(rpc.DefaultDebugPath, rpc.NewDebugHTTP(newServer))
	httpServer := &http.Server{
		Handler: handler,
	}

	err = httpServer.Serve(lis)
	if err != nil {
		log.Fatalf("Error starting http server: %v", err)
	}
}
