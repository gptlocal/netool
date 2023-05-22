package health_test

import (
	"github.com/gptlocal/netool/w/grpc/health"
	pb "github.com/gptlocal/netool/w/grpc/health/grpc_health_v1"
	"github.com/gptlocal/netool/w/grpc/internal/grpctest"
	"google.golang.org/grpc"
	"testing"
)

type s struct {
	grpctest.Tester
}

func Test(t *testing.T) {
	grpctest.RunSubTests(t, s{})
}

// Make sure the service implementation complies with the proto definition.
func (s) TestRegister(t *testing.T) {
	s := grpc.NewServer()
	pb.RegisterHealthServer(s, health.NewServer())
	s.Stop()
}
