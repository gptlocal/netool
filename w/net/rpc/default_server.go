package rpc

import (
	"io"
	"net"
)

var DefaultServer = NewServer()

func ServeConn(conn io.ReadWriteCloser) {
	DefaultServer.ServeConn(conn)
}

func ServeCodec(codec ServerCodec) {
	DefaultServer.ServeCodec(codec)
}

func ServeRequest(codec ServerCodec) error {
	return DefaultServer.ServeRequest(codec)
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func HandleHTTP() {
	DefaultServer.HandleHTTP(DefaultRPCPath, DefaultDebugPath)
}
