package rpc

import (
	"encoding/json"
	"io"
)

type jsonServerCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

func (c *jsonServerCodec) ReadRequest(r *Request, payload any) error {
	if err := c.dec.Decode(&r); err != nil {
		return err
	}
	return nil
}

func (c *jsonServerCodec) WriteResponse(r *Response, payload any) error {
	return c.enc.Encode(r)
}

func (c *jsonServerCodec) Close() error {
	return c.c.Close()
}

type jsonClientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

func (c *jsonClientCodec) WriteRequest(r *Request, payload any) error {
	return c.enc.Encode(&r)
}

func (c *jsonClientCodec) ReadResponse(r *Response, payload any) error {
	if err := c.dec.Decode(&r); err != nil {
		return err
	}
	return nil
}

func (c *jsonClientCodec) Close() error {
	return c.c.Close()
}

func NewJSONClientCodec(conn io.ReadWriteCloser) ClientCodec {
	return &jsonClientCodec{
		dec: json.NewDecoder(conn),
		enc: json.NewEncoder(conn),
		c:   conn,
	}
}

func NewJSONServerCodec(conn io.ReadWriteCloser) ServerCodec {
	return &jsonServerCodec{
		dec: json.NewDecoder(conn),
		enc: json.NewEncoder(conn),
		c:   conn,
	}
}
