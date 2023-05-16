package rpc

import (
	"io"
	"net"
)

var DefaultServer = NewServer()

func ServeConn(conn io.ReadWriteCloser) {
	DefaultServer.ServeConn(NewGobServerCodec(conn))
}

func ServeRequest(codec ServerCodec) error {
	return DefaultServer.ServeRequest(codec)
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func HandleHTTP() {
	DefaultServer.HandleHTTP(DefaultRPCPath, DefaultDebugPath)
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr any) error { return DefaultServer.Register(rcvr) }

// RegisterName is like Register but uses the provided name for the type instead of the receiver's concrete type.
func RegisterName(name string, rcvr any) error {
	return DefaultServer.RegisterName(name, rcvr)
}
