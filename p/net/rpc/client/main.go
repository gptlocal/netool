package main

import (
	"github.com/gptlocal/netool/p/net/rpc"
	"github.com/gptlocal/netool/p/net/rpc/service"
	"log"
)

func main() {
	client, err := rpc.DialHTTP("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	reply := &service.Reply{}
	err = client.Call("Arith.Add", &service.Args{7, 8}, reply)
	if err != nil {
		log.Fatalf("Error calling: %v", err)
	}
	log.Printf("Arith.Add: %v", reply)

	err = client.Call("Arith.Mul", &service.Args{7, 8}, reply)
	if err != nil {
		log.Fatalf("Error calling: %v", err)
	}
	log.Printf("Arith.Mul: %v", reply)

	err = client.Call("Arith.Div", &service.Args{16, 2}, reply)
	if err != nil {
		log.Fatalf("Error calling: %v", err)
	}
	log.Printf("Arith.Div: %v", reply)
}
