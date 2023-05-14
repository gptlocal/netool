package rpc

import (
	"encoding/json"
	"sync"
)

type Request struct {
	ServiceMethod string `json:"service_method"`
	Seq           uint64 `json:"seq"`
}

type Response struct {
	ServiceMethod string `json:"service_method"`
	Seq           uint64 `json:"seq"`
	Error         string `json:"error"`
}

var (
	reqPool sync.Pool
	resPool sync.Pool

	null = json.RawMessage("null")
)

func init() {
	reqPool.New = func() any { return new(Request) }
	resPool.New = func() any { return new(Response) }
}

type ClientCodec interface {
	WriteRequest(req *Request, payload any) error
	ReadResponse(res *Response, payload any) error

	Close() error
}

type ServerCodec interface {
	ReadRequest(req *Request, payload any) error
	WriteResponse(res *Response, payload any) error

	Close() error
}
