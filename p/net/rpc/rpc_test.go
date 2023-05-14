package rpc_test

import (
	"bytes"
	"encoding/gob"
	. "github.com/gptlocal/netool/p/net/rpc"
	"io"
	"net"
	"testing"
)

func Test_GobCodec(t *testing.T) {
	gob.Register(Args{})
	gob.Register(Reply{})

	cli, srv := net.Pipe()
	defer cli.Close()
	defer srv.Close()

	go func() {
		srvCodec := NewGobServerCodec(srv)
		req := &Request{}
		err := srvCodec.ReadRequest(req)
		if err != nil {
			return
		}
		t.Logf("ReadRequest: %v", req)

		args := req.Payload.(Args)

		reply := Reply{
			C: args.A + args.B,
		}
		srvCodec.WriteResponse(&Response{
			ServiceMethod: req.ServiceMethod,
			Seq:           req.Seq,
			Payload:       reply,
		})
	}()

	cliCodec := NewGobClientCodec(cli)
	err := cliCodec.WriteRequest(&Request{
		ServiceMethod: "Arith.Add",
		Seq:           123,
		Payload:       Args{7, 8},
	})

	if err != nil {
		t.Fatalf("WriteRequest: %s", err)
	}

	resp := &Response{}
	err = cliCodec.ReadResponse(resp)
	if err != nil {
		t.Fatalf("ReadResponse: %s", err)
	}

	reply := resp.Payload.(Reply)

	if reply.C != 15 {
		t.Fatalf("expected 15, got %d", reply.C)
	}
	t.Logf("ReadResponse: %v", reply.C)
}

func TestGobDecodeEncode(t *testing.T) {
	gob.Register(Args{})
	gob.Register(Reply{})

	r1, w1 := io.Pipe()
	defer r1.Close()

	var out bytes.Buffer
	sc := NewGobServerCodec(struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: r1,
		Writer: &out,
		Closer: io.NopCloser(nil),
	})

	go func() {
		enc := gob.NewEncoder(w1)
		err := enc.Encode(&Request{ServiceMethod: "Arith.Add", Seq: 123, Payload: Args{A: 7, B: 8}})
		if err != nil {
			t.Logf("encode err: %v", err)
		}
	}()

	r := new(Request)
	if err := sc.ReadRequest(r); err != nil {
		t.Fatal(err)
	}

	args := r.Payload.(Args)
	if err := sc.WriteResponse(&Response{
		ServiceMethod: "Arith.Add",
		Seq:           123,
		Payload:       Reply{C: args.A + args.B},
	}); err != nil {
		t.Fatal(err)
	}

	dec := gob.NewDecoder(&out)
	res := new(Response)
	if err := dec.Decode(res); err != nil {
		t.Fatal(err)
	}
	t.Logf("res: %#v", res)

	reply := res.Payload.(Reply)
	t.Logf("reply: %#v", reply)
}

func Test_GobClientCall(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	defer srv.Close()

	go RunGobService(srv)

	client := NewClient(NewGobClientCodec(cli))

	// Synchronous calls
	args := &Args{7, 8}
	reply := new(Reply)
	err := client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: got %d expected %d", reply.C, args.A+args.B)
	}

	args = &Args{7, 8}
	reply = new(Reply)
	err = client.Call("Arith.Mul", args, reply)
	if err != nil {
		t.Errorf("Mul: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A*args.B {
		t.Errorf("Mul: got %d expected %d", reply.C, args.A*args.B)
	}

	// Out of order.
	args = &Args{7, 8}
	mulReply := new(Reply)
	mulCall := client.Go("Arith.Mul", args, mulReply, nil)
	addReply := new(Reply)
	addCall := client.Go("Arith.Add", args, addReply, nil)

	addCall = <-addCall.Done
	if addCall.Error != nil {
		t.Errorf("Add: expected no error but got string %q", addCall.Error.Error())
	}
	if addReply.C != args.A+args.B {
		t.Errorf("Add: got %d expected %d", addReply.C, args.A+args.B)
	}

	mulCall = <-mulCall.Done
	if mulCall.Error != nil {
		t.Errorf("Mul: expected no error but got string %q", mulCall.Error.Error())
	}
	if mulReply.C != args.A*args.B {
		t.Errorf("Mul: got %d expected %d", mulReply.C, args.A*args.B)
	}

	// Error test
	args = &Args{7, 0}
	reply = new(Reply)
	err = client.Call("Arith.Div", args, reply)
	// expect an error: zero divide
	if err == nil {
		t.Error("Div: expected error")
	} else if err.Error() != "divide by zero" {
		t.Error("Div: expected divide by zero error; got", err)
	}
}

func Test_JSONClientCall(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	defer srv.Close()

	go RunJSONService(srv)

	client := NewClient(NewJSONClientCodec(cli))

	// Synchronous calls
	args := &Args{7, 8}
	reply := new(Reply)
	err := client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: got %d expected %d", reply.C, args.A+args.B)
	}

	args = &Args{7, 8}
	reply = new(Reply)
	err = client.Call("Arith.Mul", args, reply)
	if err != nil {
		t.Errorf("Mul: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A*args.B {
		t.Errorf("Mul: got %d expected %d", reply.C, args.A*args.B)
	}

	// Out of order.
	args = &Args{7, 8}
	mulReply := new(Reply)
	mulCall := client.Go("Arith.Mul", args, mulReply, nil)
	addReply := new(Reply)
	addCall := client.Go("Arith.Add", args, addReply, nil)

	addCall = <-addCall.Done
	if addCall.Error != nil {
		t.Errorf("Add: expected no error but got string %q", addCall.Error.Error())
	}
	if addReply.C != args.A+args.B {
		t.Errorf("Add: got %d expected %d", addReply.C, args.A+args.B)
	}

	mulCall = <-mulCall.Done
	if mulCall.Error != nil {
		t.Errorf("Mul: expected no error but got string %q", mulCall.Error.Error())
	}
	if mulReply.C != args.A*args.B {
		t.Errorf("Mul: got %d expected %d", mulReply.C, args.A*args.B)
	}

	// Error test
	args = &Args{7, 0}
	reply = new(Reply)
	err = client.Call("Arith.Div", args, reply)
	// expect an error: zero divide
	if err == nil {
		t.Error("Div: expected error")
	} else if err.Error() != "divide by zero" {
		t.Error("Div: expected divide by zero error; got", err)
	}
}
