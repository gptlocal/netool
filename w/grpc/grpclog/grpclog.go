package grpclog

import "github.com/gptlocal/netool/w/grpc/internal/grpclog"

func init() {
	SetLoggerV2(newLoggerV2())
}

// V reports whether verbosity level l is at least the requested verbose level.
func V(l int) bool {
	return grpclog.Logger.V(l)
}
