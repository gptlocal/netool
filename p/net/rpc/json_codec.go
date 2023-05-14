package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

var (
	jsonClientReqPool sync.Pool
	jsonClientResPool sync.Pool
	jsonServerReqPool sync.Pool
	jsonServerResPool sync.Pool

	null = json.RawMessage("null")
)

func init() {
	jsonServerReqPool.New = func() any { return new(serverRequest) }
	jsonServerResPool.New = func() any { return new(serverResponse) }
	jsonClientReqPool.New = func() any { return new(clientRequest) }
	jsonClientResPool.New = func() any { return new(clientResponse) }
}

type clientRequest struct {
	Method string `json:"method"`
	Params [1]any `json:"params"`
	Id     uint64 `json:"id"`
}

type clientResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  any              `json:"error"`
}

func (r *clientResponse) reset() {
	r.Id = 0
	r.Result = nil
	r.Error = nil
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}

func (r *serverRequest) reset() *serverRequest {
	r.Method = ""
	r.Params = nil
	r.Id = nil
	return r
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result any              `json:"result"`
	Error  any              `json:"error"`
}

type jsonServerCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	mutex   sync.Mutex // protects pending
	seq     uint64
	pending map[uint64]*serverRequest
}

func (c *jsonServerCodec) ReadRequestHeader(r *Request) error {
	req := jsonServerReqPool.Get().(*serverRequest).reset()
	if err := c.dec.Decode(&req); err != nil {
		return err
	}

	r.ServiceMethod = req.Method
	if err := json.Unmarshal(*req.Id, &r.Seq); err != nil {
		return err
	}

	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = req
	c.mutex.Unlock()
	return nil
}

func (c *jsonServerCodec) ReadRequestBody(x any) error {
	if x == nil {
		return nil
	}

	var params *json.RawMessage
	c.mutex.Lock()
	if req, ok := c.pending[c.seq]; ok {
		params = req.Params
	}
	c.mutex.Unlock()

	var result [1]any
	result[0] = x
	return json.Unmarshal(*params, &result)
}

func (c *jsonServerCodec) WriteResponse(r *Response, x any) error {
	resp := jsonServerResPool.Get().(*serverResponse)
	if bytes, err := json.Marshal(r.Seq); err != nil {
		return err
	} else {
		resp.Id = (*json.RawMessage)(&bytes)
	}

	if r.Error != "" {
		resp.Error = r.Error
	} else {
		resp.Result = x
	}
	return c.enc.Encode(resp)
}

func (c *jsonServerCodec) Close() error {
	return c.c.Close()
}

type jsonClientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	mutex   sync.Mutex // protects pending
	seq     uint64
	pending map[uint64]*clientResponse
}

func (c *jsonClientCodec) WriteRequest(r *Request, param any) error {
	req := jsonClientReqPool.Get().(*clientRequest)
	req.Id = r.Seq
	req.Method = r.ServiceMethod
	req.Params[0] = param
	return c.enc.Encode(&req)
}

func (c *jsonClientCodec) ReadResponseHeader(r *Response) error {
	res := jsonClientResPool.Get().(*clientResponse)
	if err := c.dec.Decode(&res); err != nil {
		return err
	}

	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = res
	c.mutex.Unlock()

	r.Seq = res.Id
	if res.Error != nil {
		r.Error = res.Error.(string)
	}
	return nil
}

func (c *jsonClientCodec) ReadResponseBody(x any) error {
	if res, ok := c.pending[c.seq]; ok {
		if res.Result == nil {
			return nil
		}
		if err := json.Unmarshal(*res.Result, x); err != nil {
			return err
		}
		jsonClientResPool.Put(res)
		return nil
	} else {
		return fmt.Errorf("invalid sequence number in response: %d", c.seq)
	}
}

func (c *jsonClientCodec) Close() error {
	return c.c.Close()
}

func NewJSONClientCodec(conn io.ReadWriteCloser) ClientCodec {
	return &jsonClientCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]*clientResponse),
	}
}

func NewJSONServerCodec(conn io.ReadWriteCloser) ServerCodec {
	return &jsonServerCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]*serverRequest),
	}
}
