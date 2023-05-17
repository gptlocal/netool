package rpc_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/gptlocal/netool/p/net/rpc"
)

var (
	newServer                 *Server
	serverAddr, newServerAddr string
	httpServerAddr            string
	newOnce                   sync.Once
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

func TestCloseCodec(t *testing.T) {
	codec := &shutdownCodec{responded: make(chan int)}
	client := NewClient(codec)
	<-codec.responded
	client.Close()
	if !codec.closed {
		t.Error("client.Close did not close codec")
	}
}

func TestGobError(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("no error")
		}
		if !strings.Contains(err.(error).Error(), "reading body unexpected EOF") {
			t.Fatal("expected `reading body unexpected EOF', got", err)
		}
	}()

	Register(new(S))

	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go Accept(listen)

	client, err := Dial("tcp", listen.Addr().String())
	if err != nil {
		panic(err)
	}

	var reply Reply
	err = client.Call("S.Recv", &struct{}{}, &reply)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", reply)

	client.Close()
	listen.Close()
}

func TestRPC(t *testing.T) {
	once.Do(startServer)
	testRPC(t, serverAddr)
	newOnce.Do(startNewServer)
	testRPC(t, newServerAddr)
	testNewServerRPC(t, newServerAddr)
}

func testRPC(t *testing.T, addr string) {
	client, err := Dial("tcp", addr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	defer client.Close()

	// Synchronous calls
	args := &Args{7, 8}
	reply := new(Reply)
	err = client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}

	// Methods exported from unexported embedded structs
	args = &Args{7, 0}
	reply = new(Reply)
	err = client.Call("Embed.Exported", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}

	// Nonexistent method
	args = &Args{7, 0}
	reply = new(Reply)
	err = client.Call("Arith.BadOperation", args, reply)
	// expect an error
	if err == nil {
		t.Error("BadOperation: expected error")
	} else if !strings.HasPrefix(err.Error(), "rpc: can't find method ") {
		t.Errorf("BadOperation: expected can't find method error; got %q", err)
	}

	// Unknown service
	args = &Args{7, 8}
	reply = new(Reply)
	err = client.Call("Arith.Unknown", args, reply)
	if err == nil {
		t.Error("expected error calling unknown service")
	} else if !strings.Contains(err.Error(), "method") {
		t.Error("expected error about method; got", err)
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
		t.Errorf("Add: expected %d got %d", addReply.C, args.A+args.B)
	}

	mulCall = <-mulCall.Done
	if mulCall.Error != nil {
		t.Errorf("Mul: expected no error but got string %q", mulCall.Error.Error())
	}
	if mulReply.C != args.A*args.B {
		t.Errorf("Mul: expected %d got %d", mulReply.C, args.A*args.B)
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

	// Bad type.
	reply = new(Reply)
	err = client.Call("Arith.Add", reply, reply) // args, reply would be the correct thing to use
	if err == nil {
		t.Error("expected error calling Arith.Add with wrong arg type")
	} else if !strings.Contains(err.Error(), "type") {
		t.Error("expected error about type; got", err)
	}

	// Non-struct argument
	const Val = 12345
	str := fmt.Sprint(Val)
	reply = new(Reply)
	err = client.Call("Arith.Scan", &str, reply)
	if err != nil {
		t.Errorf("Scan: expected no error but got string %q", err.Error())
	} else if reply.C != Val {
		t.Errorf("Scan: expected %d got %d", Val, reply.C)
	}

	// Non-struct reply
	args = &Args{27, 35}
	str = ""
	err = client.Call("Arith.String", args, &str)
	if err != nil {
		t.Errorf("String: expected no error but got string %q", err.Error())
	}
	expect := fmt.Sprintf("%d+%d=%d", args.A, args.B, args.A+args.B)
	if str != expect {
		t.Errorf("String: expected %s got %s", expect, str)
	}

	args = &Args{7, 8}
	reply = new(Reply)
	err = client.Call("Arith.Mul", args, reply)
	if err != nil {
		t.Errorf("Mul: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A*args.B {
		t.Errorf("Mul: expected %d got %d", reply.C, args.A*args.B)
	}

	// ServiceName contain "." character
	args = &Args{7, 8}
	reply = new(Reply)
	err = client.Call("net.rpc.Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}
}

func testNewServerRPC(t *testing.T, addr string) {
	client, err := Dial("tcp", addr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	defer client.Close()

	// Synchronous calls
	args := &Args{7, 8}
	reply := new(Reply)
	err = client.Call("newServer.Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}
	t.Logf("Add: %d+%d=%d", args.A, args.B, reply.C)
}

func TestHTTP(t *testing.T) {
	once.Do(startServer)
	testHTTPRPC(t, "")
	newOnce.Do(startNewServer)
	testHTTPRPC(t, newHttpPath)
}

func testHTTPRPC(t *testing.T, path string) {
	var client *Client
	var err error
	if path == "" {
		client, err = DialHTTP("tcp", httpServerAddr)
	} else {
		client, err = DialHTTPPath("tcp", httpServerAddr, path)
	}
	if err != nil {
		t.Fatal("dialing", err)
	}
	defer client.Close()

	// Synchronous calls
	args := &Args{7, 8}
	reply := new(Reply)
	err = client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}
}

func TestGobBuiltinTypes(t *testing.T) {
	once.Do(startServer)

	client, err := DialHTTP("tcp", httpServerAddr)
	if err != nil {
		t.Fatal("dialing", err)
	}
	defer client.Close()

	// Map
	args := &Args{7, 8}
	replyMap := map[int]int{}
	err = client.Call("BuiltinTypes2.Map", args, &replyMap)
	if err != nil {
		t.Errorf("Map: expected no error but got string %q", err.Error())
	}
	if replyMap[args.A] != args.B {
		t.Errorf("Map: expected %d got %d", args.B, replyMap[args.A])
	}

	// Slice
	args = &Args{7, 8}
	replySlice := []int{}
	err = client.Call("BuiltinTypes2.Slice", args, &replySlice)
	if err != nil {
		t.Errorf("Slice: expected no error but got string %q", err.Error())
	}
	if e := []int{args.A, args.B}; !reflect.DeepEqual(replySlice, e) {
		t.Errorf("Slice: expected %v got %v", e, replySlice)
	}

	// Array
	args = &Args{7, 8}
	replyArray := [2]int{}
	err = client.Call("BuiltinTypes2.Array", args, &replyArray)
	if err != nil {
		t.Errorf("Array: expected no error but got string %q", err.Error())
	}
	if e := [2]int{args.A, args.B}; !reflect.DeepEqual(replyArray, e) {
		t.Errorf("Array: expected %v got %v", e, replyArray)
	}
}

func TestServeRequest(t *testing.T) {
	once.Do(startServer)
	testServeRequest(t, nil)
	newOnce.Do(startNewServer)
	testServeRequest(t, newServer)
}

func testServeRequest(t *testing.T, server *Server) {
	client := CodecEmulator{server: server}
	defer client.Close()

	args := &Args{7, 8}
	reply := new(Reply)
	err := client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Errorf("Add: expected no error but got string %q", err.Error())
	}
	if reply.C != args.A+args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
	}
	t.Logf("Add: %d+%d=%d", args.A, args.B, reply.C)

	err = client.Call("Arith.Add", nil, reply)
	if err == nil {
		t.Errorf("expected error calling Arith.Add with nil arg")
	}
	t.Logf("Add: got expected error %q", err.Error())
}

func TestRegistrationError(t *testing.T) {
	err := Register(new(ReplyNotPointer))
	if err == nil {
		t.Error("expected error registering ReplyNotPointer")
	}
	err = Register(new(ArgNotPublic))
	if err == nil {
		t.Error("expected error registering ArgNotPublic")
	}
	err = Register(new(ReplyNotPublic))
	if err == nil {
		t.Error("expected error registering ReplyNotPublic")
	}
	err = Register(NeedsPtrType(0))
	if err == nil {
		t.Error("expected error registering NeedsPtrType")
	} else if !strings.Contains(err.Error(), "pointer") {
		t.Error("expected hint when registering NeedsPtrType")
	}
}

func TestSendDeadlock(t *testing.T) {
	client := NewClient(WriteFailCodec(0))
	defer client.Close()

	done := make(chan bool)
	go func() {
		testSendDeadlock(client)
		testSendDeadlock(client)
		done <- true
	}()
	select {
	case <-done:
		return
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock")
	}
}

func testSendDeadlock(client *Client) {
	defer func() {
		recover()
	}()
	args := &Args{7, 8}
	reply := new(Reply)
	client.Call("Arith.Add", args, reply)
}

func countMallocs(dial func() (*Client, error), t *testing.T) float64 {
	once.Do(startServer)
	client, err := dial()
	if err != nil {
		t.Fatal("error dialing", err)
	}
	defer client.Close()

	args := &Args{7, 8}
	reply := new(Reply)
	return testing.AllocsPerRun(100, func() {
		err := client.Call("Arith.Add", args, reply)
		if err != nil {
			t.Errorf("Add: expected no error but got string %q", err.Error())
		}
		if reply.C != args.A+args.B {
			t.Errorf("Add: expected %d got %d", reply.C, args.A+args.B)
		}
	})
}

func TestCountMallocs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping; GOMAXPROCS>1")
	}
	fmt.Printf("mallocs per rpc round trip: %v\n", countMallocs(dialDirect, t))
}

func TestCountMallocsOverHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}
	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping; GOMAXPROCS>1")
	}
	fmt.Printf("mallocs per HTTP rpc round trip: %v\n", countMallocs(dialHTTP, t))
}

func TestClientWriteError(t *testing.T) {
	w := &writeCrasher{done: make(chan bool)}
	c := NewClient(NewGobClientCodec(w))
	defer c.Close()

	res := false
	err := c.Call("foo", 1, &res)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "fake write failure" {
		t.Error("unexpected value of error:", err)
	}
	t.Logf("res: %v, error: %v", res, err)
	w.done <- true
}

func TestTCPClose(t *testing.T) {
	once.Do(startServer)

	client, err := dialHTTP()
	if err != nil {
		t.Fatalf("dialing: %v", err)
	}
	defer client.Close()

	args := Args{17, 8}
	var reply Reply
	err = client.Call("Arith.Mul", args, &reply)
	if err != nil {
		t.Fatal("arith error:", err)
	}
	t.Logf("Arith: %d*%d=%d\n", args.A, args.B, reply)
	if reply.C != args.A*args.B {
		t.Errorf("Add: expected %d got %d", reply.C, args.A*args.B)
	}
}

func TestErrorAfterClientClose(t *testing.T) {
	once.Do(startServer)

	client, err := dialHTTP()
	if err != nil {
		t.Fatalf("dialing: %v", err)
	}
	err = client.Close()
	if err != nil {
		t.Fatal("close error:", err)
	}
	err = client.Call("Arith.Add", &Args{7, 9}, new(Reply))
	if err != ErrShutdown {
		t.Errorf("Forever: expected ErrShutdown got %v", err)
	}
}

// Tests the fix to issue 11221. Without the fix, this loops forever or crashes.
func TestAcceptExitAfterListenerClose(t *testing.T) {
	newServer := NewServer()
	newServer.Register(new(Arith))
	newServer.RegisterName("net.rpc.Arith", new(Arith))
	newServer.RegisterName("newServer.Arith", new(Arith))

	var l net.Listener
	l, _ = listenTCP()
	l.Close()
	newServer.Accept(l)
}

func TestShutdown(t *testing.T) {
	var l net.Listener
	l, _ = listenTCP()
	ch := make(chan net.Conn, 1)
	go func() {
		defer l.Close()
		c, err := l.Accept()
		if err != nil {
			t.Error(err)
		}
		ch <- c
	}()
	c, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c1 := <-ch
	if c1 == nil {
		t.Fatal(err)
	}

	newServer := NewServer()
	newServer.Register(new(Arith))
	go newServer.ServeConn(NewGobServerCodec(c1))

	args := &Args{7, 8}
	reply := new(Reply)
	client := NewClient(NewGobClientCodec(c))
	err = client.Call("Arith.Add", args, reply)
	if err != nil {
		t.Fatal(err)
	}

	// On an unloaded system 10ms is usually enough to fail 100% of the time
	// with a broken server. On a loaded system, a broken server might incorrectly
	// be reported as passing, but we're OK with that kind of flakiness.
	// If the code is correct, this test will never fail, regardless of timeout.
	args.A = 10 // 10 ms
	done := make(chan *Call, 1)
	call := client.Go("Arith.SleepMilli", args, reply, done)
	c.(*net.TCPConn).CloseWrite()
	<-done
	if call.Error != nil {
		t.Fatal(err)
	}
}
