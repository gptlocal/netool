package service

import (
	"errors"
	"fmt"
)

type Args struct {
	A int `json:"a"`
	B int `json:"b"`
}

func (a *Args) String() string {
	return fmt.Sprintf("%d, %d", a.A, a.B)
}

type Reply struct {
	C int `json:"c"`
}

func (r *Reply) String() string {
	return fmt.Sprintf("%d", r.C)
}

type Arith int

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
