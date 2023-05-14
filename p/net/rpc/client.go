package rpc

import (
	"errors"
	"io"
	"log"
	"reflect"
	"sync"
)

type Client struct {
	codec ClientCodec

	mutex    sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

func NewClient(codec ClientCodec) *Client {
	client := &Client{
		codec:   codec,
		pending: make(map[uint64]*Call),
	}
	go client.input()
	return client
}

func (client *Client) Call(serviceMethod string, args any, reply any) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

func (client *Client) Go(serviceMethod string, args any, reply any, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(call)
	return call
}

func (client *Client) input() {
	var err error
	var response Response
	for err == nil {
		response = Response{}
		err = client.codec.ReadResponse(&response)
		if err != nil {
			break
		}
		seq := response.Seq

		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
			err = errors.New("rpc: unexpected sequence number in response")
		case response.Error != "":
			call.Error = errors.New(response.Error)
			call.done()
		default:
			reflect.ValueOf(call.Reply).Elem().Set(reflect.ValueOf(response.Payload))
			call.done()
		}
	}

	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()

	if err != io.EOF && !closing {
		log.Println("rpc: client protocol error:", err)
	}
}

func (client *Client) send(call *Call) {
	client.mutex.Lock()
	if client.shutdown || client.closing {
		client.mutex.Unlock()
		call.Error = ErrShutdown
		call.done()
		return
	}

	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	// todo: use sync.Pool
	request := &Request{
		Seq:           seq,
		ServiceMethod: call.ServiceMethod,
		Payload:       call.Args,
	}
	err := client.codec.WriteRequest(request)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		call.Error = err
		call.done()
	}
}
