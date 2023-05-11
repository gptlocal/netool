package rpc_test

import (
	"errors"
	. "github.com/gptlocal/netool/w/net/rpc"
)

type shutdownCodec struct {
	responded chan int
	closed    bool
}

func (c *shutdownCodec) WriteRequest(*Request, any) error { return nil }
func (c *shutdownCodec) ReadResponseBody(any) error       { return nil }
func (c *shutdownCodec) ReadResponseHeader(*Response) error {
	c.responded <- 1
	return errors.New("shutdownCodec ReadResponseHeader")
}
func (c *shutdownCodec) Close() error {
	c.closed = true
	return nil
}

type WriteFailCodec int

func (WriteFailCodec) WriteRequest(*Request, any) error {
	// the panic caused by this error used to not unlock a lock.
	return errors.New("fail")
}

func (WriteFailCodec) ReadResponseHeader(*Response) error {
	select {}
}

func (WriteFailCodec) ReadResponseBody(any) error {
	select {}
}

func (WriteFailCodec) Close() error {
	return nil
}
