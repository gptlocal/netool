package rpc

type Request struct {
	ServiceMethod string
	Seq           uint64
}

type Response struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

type ClientCodec interface {
	WriteRequest(*Request, any) error
	ReadResponseHeader(*Response) error
	ReadResponseBody(any) error

	Close() error
}

type ServerCodec interface {
	ReadRequestHeader(*Request) error
	ReadRequestBody(any) error
	WriteResponse(*Response, any) error

	Close() error
}
