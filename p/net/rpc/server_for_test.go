package rpc_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	. "github.com/gptlocal/netool/p/net/rpc"
	"github.com/gptlocal/netool/w/net/http/httptest"
)

const (
	newHttpPath = "/foo"
)

var (
	httpOnce sync.Once
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

type BuiltinTypes2 struct{}

func (BuiltinTypes2) Map(args *Args, reply *map[int]int) error {
	(*reply)[args.A] = args.B
	return nil
}

func (BuiltinTypes2) Slice(args *Args, reply *[]int) error {
	*reply = append(*reply, args.A, args.B)
	return nil
}

func (BuiltinTypes2) Array(args *Args, reply *[2]int) error {
	(*reply)[0] = args.A
	(*reply)[1] = args.B
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

type shutdownCodec struct {
	responded chan int
	closed    bool
}

func (c *shutdownCodec) WriteRequest(Request, any) error { return nil }
func (c *shutdownCodec) ReadResponse() (Response, error) { return &AResponse{}, nil }
func (c *shutdownCodec) ReadPayload(Response, any) error {
	c.responded <- 1
	return errors.New("shutdownCodec ReadResponseHeader")
}
func (c *shutdownCodec) Close() error {
	c.closed = true
	return nil
}

// CodecEmulator provides a client-like api and a ServerCodec interface.
// Can be used to test ServeRequest.
type CodecEmulator struct {
	server        *Server
	serviceMethod string
	args          *Args
	reply         *Reply
	err           error
}

func (codec *CodecEmulator) Call(serviceMethod string, args *Args, reply *Reply) error {
	codec.serviceMethod = serviceMethod
	codec.args = args
	codec.reply = reply
	codec.err = nil
	var serverError error
	if codec.server == nil {
		serverError = ServeRequest(codec)
	} else {
		serverError = codec.server.ServeRequest(codec)
	}
	if codec.err == nil && serverError != nil {
		codec.err = serverError
	}
	return codec.err
}

func (codec *CodecEmulator) ReadRequest() (Request, error) {
	req := &ARequest{}
	req.ServiceMethod = codec.serviceMethod
	req.Seq = 0
	return req, nil
}

func (codec *CodecEmulator) ReadPayload(req Request, argv any) error {
	if codec.args == nil {
		return io.ErrUnexpectedEOF
	}
	*(argv.(*Args)) = *codec.args
	return nil
}

func (codec *CodecEmulator) WriteResponse(resp Response, reply any) error {
	if resp.ErrorString() != "" {
		codec.err = errors.New(resp.ErrorString())
	} else {
		*codec.reply = *(reply.(*Reply))
	}
	return nil
}

func (codec *CodecEmulator) Close() error {
	return nil
}

type WriteFailCodec int

func (WriteFailCodec) WriteRequest(Request, any) error {
	// the panic caused by this error used to not unlock a lock.
	return errors.New("fail")
}

func (WriteFailCodec) ReadResponse() (Response, error) {
	select {}
}

func (WriteFailCodec) ReadPayload(Response, any) error {
	select {}
}

func (WriteFailCodec) Close() error {
	return nil
}

type ReplyNotPointer int
type ArgNotPublic int
type ReplyNotPublic int
type NeedsPtrType int
type local struct{}

func (t *ReplyNotPointer) ReplyNotPointer(args *Args, reply Reply) error {
	return nil
}

func (t *ArgNotPublic) ArgNotPublic(args *local, reply *Reply) error {
	return nil
}

func (t *ReplyNotPublic) ReplyNotPublic(args *Args, reply *local) error {
	return nil
}

func (t *NeedsPtrType) NeedsPtrType(args *Args, reply *Reply) error {
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

type hidden int

func (t *hidden) Exported(args Args, reply *Reply) error {
	reply.C = args.A + args.B
	return nil
}

type Embed struct {
	hidden
}

type writeCrasher struct {
	done chan bool
}

func (writeCrasher) Close() error {
	return nil
}

func (w *writeCrasher) Read(p []byte) (int, error) {
	<-w.done
	return 0, io.EOF
}

func (writeCrasher) Write(p []byte) (int, error) {
	return 0, errors.New("fake write failure")
}

func startServer() {
	Register(new(Arith))
	Register(new(Embed))
	RegisterName("net.rpc.Arith", new(Arith))
	Register(BuiltinTypes2{})

	var l net.Listener
	l, serverAddr = listenTCP()
	log.Println("Test RPC server listening on", serverAddr)
	go Accept(l)

	HandleHTTP()
	httpOnce.Do(startHttpServer)
}

func startNewServer() {
	newServer = NewServer()
	newServer.Register(new(Arith))
	newServer.Register(new(Embed))
	newServer.RegisterName("net.rpc.Arith", new(Arith))
	newServer.RegisterName("newServer.Arith", new(Arith))

	var l net.Listener
	l, newServerAddr = listenTCP()
	log.Println("NewServer test RPC server listening on", newServerAddr)
	go newServer.Accept(l)

	newServer.HandleHTTP(newHttpPath, "/bar")
	httpOnce.Do(startHttpServer)
}

func listenTCP() (net.Listener, string) {
	l, e := net.Listen("tcp", "127.0.0.1:0") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}
	return l, l.Addr().String()
}

func startHttpServer() {
	server := httptest.NewServer(nil)
	httpServerAddr = server.Listener.Addr().String()
	log.Println("Test HTTP RPC server listening on", httpServerAddr)
}
