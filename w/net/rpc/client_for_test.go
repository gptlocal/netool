package rpc_test

import (
	. "github.com/gptlocal/netool/w/net/rpc"
)

func dialDirect() (*Client, error) {
	return Dial("tcp", serverAddr)
}

func dialHTTP() (*Client, error) {
	return DialHTTP("tcp", httpServerAddr)
}
