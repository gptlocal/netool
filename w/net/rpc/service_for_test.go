package rpc

import (
	"errors"
	"fmt"
	"time"
)

type (
	Args struct {
		A, B int
	}

	Reply struct {
		C int
	}

	Arith int // arithmetic

	hidden int

	Embed struct {
		hidden
	}
)

func (a *Args) String() string {
	return fmt.Sprintf("%d, %d", a.A, a.B)
}

func (r *Reply) String() string {
	return fmt.Sprintf("%d", r.C)
}

// Some of Arith's methods have value args, some have pointer args. That's deliberate.
func (t *Arith) Add(args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

func (t *Arith) Mul(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *Arith) Div(args Args, reply *Reply) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	reply.C = args.A / args.B
	return nil
}

func (t *Arith) String(args *Args, reply *string) error {
	*reply = fmt.Sprintf("%d+%d=%d", args.A, args.B, args.A+args.B)
	return nil
}

func (t *Arith) Scan(args string, reply *Reply) (err error) {
	_, err = fmt.Sscan(args, &reply.C)
	return
}

func (t *Arith) Error(args *Args, reply *Reply) error {
	panic("ERROR")
}

func (t *Arith) SleepMilli(args *Args, reply *Reply) error {
	time.Sleep(time.Duration(args.A) * time.Millisecond)
	return nil
}

func (t *hidden) Exported(args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type BuiltinTypes struct{}

func (BuiltinTypes) Map(args *Args, reply *map[int]int) error {
	(*reply)[args.A] = args.B
	return nil
}

func (BuiltinTypes) Slice(args *Args, reply *[]int) error {
	*reply = append(*reply, args.A, args.B)
	return nil
}

func (BuiltinTypes) Array(args *Args, reply *[2]int) error {
	(*reply)[0] = args.A
	(*reply)[1] = args.B
	return nil
}

// Test that errors in gob shut down the connection. Issue 7689.
type R struct {
	msg []byte // Not exported, so R does not work with gob.
}

type S struct{}

func (s *S) Recv(nul *struct{}, reply *R) error {
	*reply = R{[]byte("foo")}
	return nil
}
