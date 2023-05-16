package rpc

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
)

type clientRequest struct {
	ARequest
	Payload any `json:"payload"`
}

type clientResponse struct {
	AResponse
	Payload *json.RawMessage `json:"payload"`
}

type serverRequest struct {
	ARequest
	Payload *json.RawMessage `json:"payload"`
}

type serverResponse struct {
	AResponse
	Payload any `json:"payload"`
}

func (r *serverResponse) reset() {
	r.Error = ""
	r.Payload = nil
}

var (
	clientReqPool sync.Pool
	clientResPool sync.Pool
	serverReqPool sync.Pool
	serverResPool sync.Pool
)

func init() {
	clientReqPool.New = func() any { return new(clientRequest) }
	clientResPool.New = func() any { return new(clientResponse) }
	serverReqPool.New = func() any { return new(serverRequest) }
	serverResPool.New = func() any { return new(serverResponse) }
}

type jsonServerCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

func (c *jsonServerCodec) ReadRequest() (Request, error) {
	jsonReq := serverReqPool.Get().(*serverRequest)
	if err := c.dec.Decode(&jsonReq); err != nil {
		serverReqPool.Put(jsonReq)
		return nil, err
	}
	return jsonReq, nil
}

func (c *jsonServerCodec) ReadPayload(req Request, payload any) error {
	if jsonReq, ok := req.(*serverRequest); ok {
		defer serverReqPool.Put(jsonReq)
		if jsonReq.Payload == nil {
			return errMissingParams
		}
		if payload != nil {
			return json.Unmarshal(*jsonReq.Payload, &payload)
		}
		return nil
	}
	return errors.New("not a json request")
}

func (c *jsonServerCodec) WriteResponse(r Response, payload any) error {
	jsonRes := serverResPool.Get().(*serverResponse)
	defer func() {
		jsonRes.reset()
		serverResPool.Put(jsonRes)
	}()

	jsonRes.ServiceMethod = r.Method()
	jsonRes.Seq = r.SeqId()
	if r.ErrorString() != "" {
		jsonRes.Error = r.ErrorString()
	} else {
		jsonRes.Payload = payload
	}
	return c.enc.Encode(&jsonRes)
}

func (c *jsonServerCodec) Close() error {
	return c.c.Close()
}

type jsonClientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

func (c *jsonClientCodec) WriteRequest(r Request, payload any) error {
	jsonReq := clientReqPool.Get().(*clientRequest)
	defer clientReqPool.Put(jsonReq)

	jsonReq.ServiceMethod = r.Method()
	jsonReq.Seq = r.SeqId()
	jsonReq.Payload = payload
	return c.enc.Encode(&jsonReq)
}

func (c *jsonClientCodec) ReadResponse() (Response, error) {
	jsonRes := clientResPool.Get().(*clientResponse)
	if err := c.dec.Decode(&jsonRes); err != nil {
		clientResPool.Put(jsonRes)
		return nil, err
	}

	return jsonRes, nil
}

func (c *jsonClientCodec) ReadPayload(r Response, payload any) error {
	if jsonRes, ok := r.(*clientResponse); ok {
		defer clientResPool.Put(jsonRes)
		if jsonRes.Payload == nil {
			return errMissingParams
		}
		if payload != nil {
			return json.Unmarshal(*jsonRes.Payload, &payload)
		}
		return nil
	}
	return errors.New("not a json response")
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
