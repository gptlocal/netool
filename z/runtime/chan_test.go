package runtime_test

import (
	. "github.com/gptlocal/netool/z/runtime"
	"testing"
)

func Test_IsClosedChan(t *testing.T) {
	if IsClosedChan(nil) {
		t.Fatalf("IsClosedChan(c) = true, want false")
	}

	c := make(chan struct{})

	if IsClosedChan(c) {
		t.Fatalf("IsClosedChan(c) = true, want false")
	}

	s := make(chan int)
	go func() {
		s <- 1
		c <- struct{}{}
	}()

	<-s
	if IsClosedChan(c) {
		t.Fatalf("IsClosedChan(c) = true, want false")
	}

	close(c)
	if !IsClosedChan(c) {
		t.Fatalf("IsClosedChan(c) = false, want true")
	}
}
