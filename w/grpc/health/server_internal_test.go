package health

import (
	pb "github.com/gptlocal/netool/w/grpc/health/grpc_health_v1"
	"github.com/gptlocal/netool/w/grpc/internal/grpctest"
	"sync"
	"testing"
	"time"
)

type s struct {
	grpctest.Tester
}

func Test(t *testing.T) {
	grpctest.RunSubTests(t, s{})
}

func (s) TestShutdown(t *testing.T) {
	const testService = "tteesstt"
	s := NewServer()
	s.SetServingStatus(testService, pb.HealthCheckResponse_SERVING)

	status := s.statusMap[testService]
	if status != pb.HealthCheckResponse_SERVING {
		t.Fatalf("status for %s is %v, want %v", testService, status, pb.HealthCheckResponse_SERVING)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// Run SetServingStatus and Shutdown in parallel.
	go func() {
		for i := 0; i < 1000; i++ {
			s.SetServingStatus(testService, pb.HealthCheckResponse_SERVING)
			time.Sleep(time.Microsecond)
		}
		wg.Done()
	}()
	go func() {
		time.Sleep(300 * time.Microsecond)
		s.Shutdown()
		wg.Done()
	}()
	wg.Wait()

	s.mu.Lock()
	status = s.statusMap[testService]
	s.mu.Unlock()
	if status != pb.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("status for %s is %v, want %v", testService, status, pb.HealthCheckResponse_NOT_SERVING)
	}

	s.Resume()
	status = s.statusMap[testService]
	if status != pb.HealthCheckResponse_SERVING {
		t.Fatalf("status for %s is %v, want %v", testService, status, pb.HealthCheckResponse_SERVING)
	}

	s.SetServingStatus(testService, pb.HealthCheckResponse_NOT_SERVING)
	status = s.statusMap[testService]
	if status != pb.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("status for %s is %v, want %v", testService, status, pb.HealthCheckResponse_NOT_SERVING)
	}
}
