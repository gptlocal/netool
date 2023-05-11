package rpc

import (
	"fmt"
	"log"
)

type Call struct {
	ServiceMethod string     // The name of the service and method to call.
	Args          any        // The argument to the function (*struct).
	Reply         any        // The reply from the function (*struct).
	Error         error      // After completion, the error status.
	Done          chan *Call // Receives *Call when Go is complete.
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		log.Printf("rpc: call done: %v", call)
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		if debugLog {
			log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}

func (call *Call) String() string {
	return fmt.Sprintf("Call %s(%v) (%v, %v)", call.ServiceMethod, call.Args, call.Reply, call.Error)
}
