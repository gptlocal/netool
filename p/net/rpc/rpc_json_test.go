package rpc_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	. "github.com/gptlocal/netool/p/net/rpc"
)

func TestServerNoParams(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	go RunJSONService(srv)
	dec := json.NewDecoder(cli)

	fmt.Fprintf(cli, `{"method": "Arith.Add", "seq": 987}`)
	var resp ArithResp
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("Decode after no params: %s", err)
	}
	t.Logf("resp after no params: %#v", resp)

	if resp.Error == "" {
		t.Fatalf("Expected error, got nil")
	}
}

func TestServerEmptyMessage(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	go RunJSONService(srv)
	dec := json.NewDecoder(cli)

	fmt.Fprintf(cli, "{}")
	var resp ArithResp
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("Decode after empty: %s", err)
	}
	t.Logf("resp after no params: %#v", resp)

	if resp.Error == "" {
		t.Fatalf("Expected error, got nil")
	}
}

func TestJSONServer(t *testing.T) {
	cli, srv := net.Pipe()
	defer cli.Close()
	go RunJSONService(srv)
	dec := json.NewDecoder(cli)

	// Send hand-coded requests to server, parse responses.
	for i := 0; i < 10; i++ {
		fmt.Fprintf(cli, `{"method": "Arith.Add", "seq": %d, "payload": {"A": %d, "B": %d}}`, i, i, i+1)
		var resp ArithResp
		err := dec.Decode(&resp)
		if err != nil {
			t.Fatalf("Decode: %s", err)
		}

		t.Logf("resp: %#v", resp)

		if resp.Error != "" {
			t.Fatalf("resp.Error: %s", resp.Error)
		}

		if resp.Id != i {
			t.Fatalf("resp: bad id %d want %d", resp.Result.C, i)
		}

		if resp.Result.C != 2*i+1 {
			t.Fatalf("resp: bad result: %d+%d=%d", i, i+1, resp.Result.C)
		}
	}
}

func TestJSONClient(t *testing.T) {
	// Assume server is okay (TestServer is above).
	// Test client against server.
	cli, srv := net.Pipe()
	go RunJSONService(srv)

	client := NewClient(NewJSONClientCodec(cli))
	defer cli.Close()

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

func TestJSONBuiltinTypes(t *testing.T) {
	cli, srv := net.Pipe()
	go RunJSONService(srv)

	client := NewClient(NewJSONClientCodec(cli))
	defer cli.Close()

	// Map
	arg := 7
	replyMap := map[int]int{}
	err := client.Call("BuiltinTypes.Map", arg, &replyMap)
	if err != nil {
		t.Errorf("Map: expected no error but got string %q", err.Error())
	}
	if replyMap[arg] != arg {
		t.Errorf("Map: expected %d got %d", arg, replyMap[arg])
	}

	// Slice
	replySlice := []int{}
	err = client.Call("BuiltinTypes.Slice", arg, &replySlice)
	if err != nil {
		t.Errorf("Slice: expected no error but got string %q", err.Error())
	}
	if e := []int{arg}; !reflect.DeepEqual(replySlice, e) {
		t.Errorf("Slice: expected %v got %v", e, replySlice)
	}

	// Array
	replyArray := [1]int{}
	err = client.Call("BuiltinTypes.Array", arg, &replyArray)
	if err != nil {
		t.Errorf("Array: expected no error but got string %q", err.Error())
	}
	if e := [1]int{arg}; !reflect.DeepEqual(replyArray, e) {
		t.Errorf("Array: expected %v got %v", e, replyArray)
	}
}

func TestJSONMalformedInput(t *testing.T) {
	cli, srv := net.Pipe()
	go cli.Write([]byte(`{id:1}`)) // invalid json
	RunJSONService(srv)            // must return, not loop
}

func TestMalformedOutput(t *testing.T) {
	cli, srv := net.Pipe()
	go srv.Write([]byte(`{"seq":0,"payload":null,"error":""}`))
	go io.ReadAll(srv)

	client := NewClient(NewJSONClientCodec(cli))
	defer cli.Close()

	args := &Args{7, 8}
	reply := new(Reply)
	err := client.Call("Arith.Add", args, reply)
	if err == nil {
		t.Error("expected error")
	}
	t.Logf("expected error: %s, reply: %v", err, reply)
}

func TestServerErrorHasNullResult(t *testing.T) {
	var out strings.Builder
	sc := NewJSONServerCodec(struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: strings.NewReader(`{"method": "Arith.Add", "seq": 123}`),
		Writer: &out,
		Closer: io.NopCloser(nil),
	})

	_, err := sc.ReadRequest()
	if err != nil {
		t.Fatal(err)
	}

	const valueText = "the value we don't want to see"
	const errorText = "some error"
	err = sc.WriteResponse(&AResponse{
		ServiceMethod: "Method",
		Seq:           1,
		Error:         errorText,
	}, valueText)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("response: %s", &out)

	if !strings.Contains(out.String(), errorText) {
		t.Fatalf("Response didn't contain expected error %q: %s", errorText, &out)
	}
	if strings.Contains(out.String(), valueText) {
		t.Errorf("Response contains both an error and value: %s", &out)
	}
}

func TestUnexpectedError(t *testing.T) {
	cli, srv := myPipe()
	go cli.PipeWriter.CloseWithError(errors.New("unexpected error!")) // reader will get this error
	RunJSONService(srv)                                               // must return, not loop
}
