package grpctest

import (
	"github.com/gptlocal/netool/w/grpc/internal/leakcheck"
	"sync/atomic"
	"testing"
)

var lcFailed uint32

type errorer struct {
	t *testing.T
}

func (e errorer) Errorf(format string, args ...interface{}) {
	atomic.StoreUint32(&lcFailed, 1)
	e.t.Errorf(format, args...)
}

// Tester is an implementation of the x interface parameter to grpctest.RunSubTests with default Setup and Teardown
// behavior. Setup updates the tlogger and Teardown performs a leak check. Embed in a struct with tests defined to use.
type Tester struct{}

// Setup updates the tlogger.
func (Tester) Setup(t *testing.T) {
	TLogger.Update(t)
}

// Teardown performs a leak check.
func (Tester) Teardown(t *testing.T) {
	if atomic.LoadUint32(&lcFailed) == 1 {
		return
	}
	leakcheck.Check(errorer{t: t})
	if atomic.LoadUint32(&lcFailed) == 1 {
		t.Log("Leak check disabled for future tests")
	}
	TLogger.EndTest(t)
}
