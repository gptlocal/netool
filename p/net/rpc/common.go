package rpc

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync"
)

type Call struct {
	ServiceMethod string
	Args          any
	Reply         any
	Error         error
	Done          chan *Call
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		log.Printf("rpc: call done: %v", call)
	default:
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func (call *Call) String() string {
	return fmt.Sprintf("Call %s(%v) (%v, %v)", call.ServiceMethod, call.Args, call.Reply, call.Error)
}

type methodType struct {
	sync.Mutex
	numCalls  uint
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

type service struct {
	name   string
	rcvr   reflect.Value
	typ    reflect.Type
	method map[string]*methodType
}

func (s *service) call(server *Server, sending *sync.Mutex, wg *sync.WaitGroup, mtype *methodType, resp *AResponse, argv, replyv reflect.Value, codec ServerCodec) {
	if wg != nil {
		defer wg.Done()
	}
	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	if errInter != nil {
		resp.Error = errInter.(error).Error()
	}
	server.sendResponse(sending, codec, resp, replyv.Interface())
}

type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var (
	ErrShutdown = errors.New("connection is shut down")

	typeOfError = reflect.TypeOf((*error)(nil)).Elem()

	errMissingParams = errors.New("jsonrpc: request body missing params")
)
