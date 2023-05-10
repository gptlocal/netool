package rpc_test

import (
	"log"
	"net"
	"sync"
	"testing"

	"github.com/gptlocal/netool/w/net/http/httptest"
	. "github.com/gptlocal/netool/w/net/rpc"
)

const (
	newHttpPath = "/foo"
)

var (
	newServer                 *Server
	serverAddr, newServerAddr string
	httpServerAddr            string
	once, newOnce, httpOnce   sync.Once
)

func TestRPC(t *testing.T) {
	once.Do(startServer)
	//testRPC(t, serverAddr)
	newOnce.Do(startNewServer)
	//testRPC(t, newServerAddr)
	//testNewServerRPC(t, newServerAddr)
}

func listenTCP() (net.Listener, string) {
	l, e := net.Listen("tcp", "127.0.0.1:0") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}
	return l, l.Addr().String()
}

func startServer() {
	//Register(new(Arith))
	//Register(new(Embed))
	//RegisterName("net.rpc.Arith", new(Arith))
	//Register(BuiltinTypes{})

	var l net.Listener
	l, serverAddr = listenTCP()
	log.Println("Test RPC server listening on", serverAddr)
	go Accept(l)

	HandleHTTP()
	httpOnce.Do(startHttpServer)
}

func startNewServer() {
	newServer = NewServer()
	//newServer.Register(new(Arith))
	//newServer.Register(new(Embed))
	//newServer.RegisterName("net.rpc.Arith", new(Arith))
	//newServer.RegisterName("newServer.Arith", new(Arith))

	var l net.Listener
	l, newServerAddr = listenTCP()
	log.Println("NewServer test RPC server listening on", newServerAddr)
	go newServer.Accept(l)

	newServer.HandleHTTP(newHttpPath, "/bar")
	httpOnce.Do(startHttpServer)
}

func startHttpServer() {
	server := httptest.NewServer(nil)
	httpServerAddr = server.Listener.Addr().String()
	log.Println("Test HTTP RPC server listening on", httpServerAddr)
}
