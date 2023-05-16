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
		req, err := srvCodec.ReadRequest()
		if err != nil {
			return
		}
		t.Logf("ReadRequest: %v", req)

		var args Args
		srvCodec.ReadPayload(req, &args)

		reply := Reply{
			C: args.A + args.B,
		}
		srvCodec.WriteResponse(&AResponse{
			ServiceMethod: req.Method(),
			Seq:           req.SeqId(),
		}, reply)
	}()

	cliCodec := NewGobClientCodec(cli)
	err := cliCodec.WriteRequest(&ARequest{
		ServiceMethod: "Arith.Add",
		Seq:           123,
	}, &Args{7, 8})

	if err != nil {
		t.Fatalf("WriteRequest: %s", err)
	}

	resp, err := cliCodec.ReadResponse()
	if err != nil {
		t.Fatalf("ReadResponse: %s", err)
	}

	var reply Reply
	cliCodec.ReadPayload(resp, &reply)

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
		if err := enc.Encode(&ARequest{ServiceMethod: "Arith.Add", Seq: 123}); err != nil {
			t.Logf("encode err: %v", err)
		}
		if err := enc.Encode(&Args{7, 9}); err != nil {
			t.Logf("encode err: %v", err)
		}
	}()

	req, err := sc.ReadRequest()
	if err != nil {
		t.Fatal(err)
	}

	var args Args
	sc.ReadPayload(req, &args)
	if err := sc.WriteResponse(&AResponse{req.Method(), req.SeqId(), ""}, &Reply{args.A + args.B}); err != nil {
		t.Fatal(err)
	}

	dec := gob.NewDecoder(&out)
	res := new(AResponse)
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
