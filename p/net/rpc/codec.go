package rpc

import (
	"encoding/json"
	"sync"
)

type Request interface {
	Method() string
	SeqId() uint64
}

type Response interface {
	Method() string
	SeqId() uint64
	ErrorString() string
}

type ARequest struct {
	ServiceMethod string `json:"method"`
	Seq           uint64 `json:"seq"`
}

func (r *ARequest) Method() string {
	return r.ServiceMethod
}

func (r *ARequest) SeqId() uint64 {
	return r.Seq
}

type AResponse struct {
	ServiceMethod string `json:"method"`
	Seq           uint64 `json:"seq"`
	Error         string `json:"error"`
}

func (r *AResponse) Method() string {
	return r.ServiceMethod
}

func (r *AResponse) SeqId() uint64 {
	return r.Seq
}

func (r *AResponse) ErrorString() string {
	return r.Error
}

var (
	reqPool sync.Pool
	resPool sync.Pool

	null = json.RawMessage("null")
)

func init() {
	reqPool.New = func() any { return new(ARequest) }
	resPool.New = func() any { return new(AResponse) }
}

type ClientCodec interface {
	WriteRequest(req Request, payload any) error
	ReadResponse() (Response, error)
	ReadPayload(res Response, payload any) error

	Close() error
}

type ServerCodec interface {
	ReadRequest() (Request, error)
	ReadPayload(req Request, payload any) error
	WriteResponse(res Response, payload any) error

	Close() error
}
