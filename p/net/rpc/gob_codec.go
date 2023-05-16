package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

var invalidRequest = struct{}{}

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r Request, payload any) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(payload); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponse() (Response, error) {
	r := new(AResponse)
	if err := c.dec.Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *gobClientCodec) ReadPayload(r Response, body any) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *gobServerCodec) ReadRequest() (Request, error) {
	r := new(ARequest)
	if err := c.dec.Decode(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *gobServerCodec) ReadPayload(r Request, payload any) error {
	return c.dec.Decode(payload)
}

func (c *gobServerCodec) WriteResponse(r Response, payload any) (err error) {
	if err = c.enc.Encode(r); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}

	if payload == nil {
		payload = invalidRequest
	}
	if err = c.enc.Encode(payload); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

func NewGobClientCodec(conn io.ReadWriteCloser) ClientCodec {
	encBuf := bufio.NewWriter(conn)
	return &gobClientCodec{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(encBuf),
		encBuf: encBuf,
	}
}

func NewGobServerCodec(conn io.ReadWriteCloser) ServerCodec {
	encBuf := bufio.NewWriter(conn)
	return &gobServerCodec{
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
		enc:    gob.NewEncoder(encBuf),
		encBuf: encBuf,
	}
}
