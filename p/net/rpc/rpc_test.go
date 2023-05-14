package rpc_test

import (
	"bytes"
	"encoding/gob"
	"io"
	"net"
	"testing"

	. "github.com/gptlocal/netool/p/net/rpc"
)

func Test_GobCodec(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	defer srv.Close()

	go func() {
		srvCodec := NewGobServerCodec(srv)
		req := &Request{}
		err := srvCodec.ReadRequestHeader(req)
		if err != nil {
			return
		}
		t.Logf("ReadRequestHeader: %v", req)

		args := &Args{}
		srvCodec.ReadRequestBody(args)
		t.Logf("ReadRequestBody: %v", args)

		reply := &Reply{
			C: args.A + args.B,
		}
		srvCodec.WriteResponse(&Response{
			ServiceMethod: req.ServiceMethod,
			Seq:           req.Seq,
		}, reply)
	}()

	cliCodec := NewGobClientCodec(cli)
	err := cliCodec.WriteRequest(&Request{
		ServiceMethod: "Arith.Add",
		Seq:           123,
	}, &Args{7, 8})

	if err != nil {
		t.Fatalf("WriteRequest: %s", err)
	}

	resp := &Response{}
	err = cliCodec.ReadResponseHeader(resp)
	if err != nil {
		t.Fatalf("ReadResponseHeader: %s", err)
	}

	reply := &Reply{}
	err = cliCodec.ReadResponseBody(reply)
	if err != nil {
		t.Fatalf("ReadResponseBody: %s", err)
	}

	if reply.C != 15 {
		t.Fatalf("expected 15, got %d", reply.C)
	}
}

func TestGobDecodeEncode(t *testing.T) {
	r1, w1 := io.Pipe()

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
		enc.Encode(&Request{ServiceMethod: "Arith.Add", Seq: 123})
		enc.Encode(&Args{A: 7, B: 8})
	}()

	r := new(Request)
	if err := sc.ReadRequestHeader(r); err != nil {
		t.Fatal(err)
	}

	args := new(Args)
	if err := sc.ReadRequestBody(args); err != nil {
		t.Fatal(err)
	}

	if err := sc.WriteResponse(&Response{ServiceMethod: "Arith.Add", Seq: 123}, &Reply{C: args.A + args.B}); err != nil {
		t.Fatal(err)
	}

	dec := gob.NewDecoder(&out)
	res := new(Response)
	if err := dec.Decode(res); err != nil {
		t.Fatal(err)
	}
	t.Logf("res: %#v", res)

	reply := new(Reply)
	if err := dec.Decode(reply); err != nil {
		t.Fatal(err)
	}
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
