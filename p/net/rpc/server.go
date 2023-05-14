package rpc

import (
	"errors"
	"fmt"
	"go/token"
	"io"
	"log"
	"reflect"
	"strings"
	"sync"
)

type Server struct {
	serviceMap sync.Map // map[string]*service
}

func NewServer() *Server {
	return &Server{}
}

func (server *Server) ServeConn(codec ServerCodec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(codec)
		if err != nil {
			log.Printf("rpc: server cannot decode request: %v", err)
			server.sendResponse(sending, req, nil, codec, err.Error())
			// todo: use sync.Pool for req
			continue
		}

		svc, mtype, err := server.findService(req)
		if err != nil {
			log.Printf("rpc: can't find service %s", req.ServiceMethod)
			server.sendResponse(sending, req, nil, codec, err.Error())
			// todo: use sync.Pool for req
			continue
		}

		wg.Add(1)

		// Decode the argument value.
		var argv reflect.Value
		argIsValue := false // if true, need to indirect before calling.
		if mtype.ArgType.Kind() == reflect.Pointer {
			argv = reflect.New(mtype.ArgType.Elem())
		} else {
			argv = reflect.New(mtype.ArgType)
			argIsValue = true
		}

		// argv guaranteed to be a pointer now.
		argv.Elem().Set(reflect.ValueOf(req.Payload))
		if argIsValue {
			argv = argv.Elem()
		}

		replyv := reflect.New(mtype.ReplyType.Elem())
		switch mtype.ReplyType.Elem().Kind() {
		case reflect.Map:
			replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
		case reflect.Slice:
			replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
		}

		go svc.call(server, sending, wg, mtype, req, argv, replyv, codec)
	}
	// We've seen that there are no more requests. Wait for responses to be sent before closing codec.
	wg.Wait()
	codec.Close()
}

func (server *Server) readRequest(codec ServerCodec) (*Request, error) {
	//todo: use sync.Pool
	req := &Request{}
	err := codec.ReadRequest(req)
	if err != nil {
		if err != io.EOF {
			log.Println("rpc:", err)
			return req, nil
		}
		return nil, err
	}
	return req, nil
}

func (server *Server) findService(req *Request) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/method request ill-formed: " + req.ServiceMethod)
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]

	// Look up the request.
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc: can't find service " + req.ServiceMethod)
	}

	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.ServiceMethod)
	}

	return svc, mtype, err
}

func (server *Server) sendResponse(sending *sync.Mutex, req *Request, reply any, codec ServerCodec, errmsg string) {
	// todo: use sync.Pool
	resp := &Response{}
	// Encode the response header
	resp.ServiceMethod = req.ServiceMethod
	if errmsg != "" {
		resp.Error = errmsg
		reply = nil
	}
	resp.Seq = req.Seq
	resp.Payload = reply
	sending.Lock()
	err := codec.WriteResponse(resp)
	if err != nil {
		log.Println("rpc: writing response:", err)
	}
	sending.Unlock()
	// todo: use sync.Pool for resp
}

func (server *Server) Register(rcvr any) error {
	return server.register(rcvr, "")
}

func (server *Server) RegisterName(name string, rcvr any) error {
	return server.register(rcvr, name)
}

func (server *Server) register(rcvr any, name string) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname, err := sanitizeServiceName(rcvr, name)
	if err != nil {
		return err
	}
	s.name = sname

	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		method := suitableMethods(reflect.PointerTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}

	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

func sanitizeServiceName(rcvr any, name string) (string, error) {
	useName := strings.TrimSpace(name) != ""
	if !useName {
		name = reflect.Indirect(reflect.ValueOf(rcvr)).Type().Name()
	}

	sname := name
	if sname == "" {
		s := fmt.Sprintf("rpc.Register: no service name for svc %v", rcvr)
		log.Printf(s)
		return sname, errors.New(s)
	}

	if !useName && !token.IsExported(sname) {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return sname, errors.New(s)
	}

	return sname, nil
}

func suitableMethods(typ reflect.Type, logErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if !method.IsExported() {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if logErr {
				log.Printf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
			}
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if logErr {
				log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Pointer {
			if logErr {
				log.Printf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if logErr {
				log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if logErr {
				log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if logErr {
				log.Printf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}
