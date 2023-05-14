package rpc_test

import (
	"errors"
	"io"

	. "github.com/gptlocal/netool/p/net/rpc"
)

type Args struct {
	A, B int
}

type Reply struct {
	C int
}

type Arith int

type ArithAddResp struct {
	Id     any   `json:"id"`
	Result Reply `json:"result"`
	Error  any   `json:"error"`
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

func RunGobService(conn io.ReadWriteCloser) {
	server := NewServer()
	server.Register(new(Arith))
	server.ServeConn(NewGobServerCodec(conn))
}

func RunJSONService(conn io.ReadWriteCloser) {
	server := NewServer()
	server.Register(new(Arith))
	server.ServeConn(NewJSONServerCodec(conn))
}
