package rpc_test

import (
	"errors"
	"io"
	"net"

	. "github.com/gptlocal/netool/p/net/rpc"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Reply struct {
	C int `json:"c"`
}

type Arith int

type ArithResp struct {
	Id     int    `json:"seq"`
	Method string `json:"method"`
	Result Reply  `json:"payload"`
	Error  string `json:"error"`
}

func (t *Arith) Add(args *Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

func (t *Arith) Mul(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *Arith) Div(args *Args, reply *Reply) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	reply.C = args.A / args.B
	return nil
}

func (t *Arith) Error(args *Args, reply *Reply) error {
	panic("ERROR")
}

type BuiltinTypes struct{}

func (BuiltinTypes) Map(i int, reply *map[int]int) error {
	(*reply)[i] = i
	return nil
}

func (BuiltinTypes) Slice(i int, reply *[]int) error {
	*reply = append(*reply, i)
	return nil
}

func (BuiltinTypes) Array(i int, reply *[1]int) error {
	(*reply)[0] = i
	return nil
}

// Copied from package net.
func myPipe() (*pipe, *pipe) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &pipe{r1, w2}, &pipe{r2, w1}
}

type pipe struct {
	*io.PipeReader
	*io.PipeWriter
}

type pipeAddr int

func (pipeAddr) Network() string {
	return "pipe"
}

func (pipeAddr) String() string {
	return "pipe"
}

func (p *pipe) Close() error {
	err := p.PipeReader.Close()
	err1 := p.PipeWriter.Close()
	if err == nil {
		err = err1
	}
	return err
}

func (p *pipe) LocalAddr() net.Addr {
	return pipeAddr(0)
}

func (p *pipe) RemoteAddr() net.Addr {
	return pipeAddr(0)
}

func (p *pipe) SetTimeout(nsec int64) error {
	return errors.New("net.Pipe does not support timeouts")
}

func (p *pipe) SetReadTimeout(nsec int64) error {
	return errors.New("net.Pipe does not support timeouts")
}

func (p *pipe) SetWriteTimeout(nsec int64) error {
	return errors.New("net.Pipe does not support timeouts")
}

func RunGobService(conn io.ReadWriteCloser) {
	server := NewServer()
	server.Register(new(Arith))
	server.ServeConn(NewGobServerCodec(conn))
}

func RunJSONService(conn io.ReadWriteCloser) {
	server := NewServer()
	server.Register(new(Arith))
	server.Register(BuiltinTypes{})
	server.ServeConn(NewJSONServerCodec(conn))
}
