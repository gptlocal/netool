package rpc_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/gptlocal/netool/p/net/rpc"
)

var (
	once sync.Once
)

func BenchmarkEndToEnd(b *testing.B) {
	benchmarkEndToEnd(dialDirect, b)
}

func BenchmarkEndToEndHTTP(b *testing.B) {
	benchmarkEndToEnd(dialHTTP, b)
}

func BenchmarkEndToEndAsync(b *testing.B) {
	benchmarkEndToEndAsync(dialDirect, b)
}

func BenchmarkEndToEndAsyncHTTP(b *testing.B) {
	benchmarkEndToEndAsync(dialHTTP, b)
}

func benchmarkEndToEnd(dial func() (*Client, error), b *testing.B) {
	once.Do(startServer)
	client, err := dial()
	if err != nil {
		b.Fatal("error dialing:", err)
	}
	defer client.Close()

	// Synchronous calls
	args := &Args{7, 8}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		reply := new(Reply)
		for pb.Next() {
			err := client.Call("Arith.Add", args, reply)
			if err != nil {
				b.Fatalf("rpc error: Add: expected no error but got string %q", err.Error())
			}
			if reply.C != args.A+args.B {
				b.Fatalf("rpc error: Add: expected %d got %d", reply.C, args.A+args.B)
			}
		}
	})
}

func benchmarkEndToEndAsync(dial func() (*Client, error), b *testing.B) {
	if b.N == 0 {
		return
	}
	const MaxConcurrentCalls = 100
	once.Do(startServer)
	client, err := dial()
	if err != nil {
		b.Fatal("error dialing:", err)
	}
	defer client.Close()

	// Asynchronous calls
	args := &Args{7, 8}
	procs := 4 * runtime.GOMAXPROCS(-1)
	send := int32(b.N)
	recv := int32(b.N)
	var wg sync.WaitGroup
	wg.Add(procs)
	gate := make(chan bool, MaxConcurrentCalls)
	res := make(chan *Call, MaxConcurrentCalls)
	b.ResetTimer()

	for p := 0; p < procs; p++ {
		go func() {
			for atomic.AddInt32(&send, -1) >= 0 {
				gate <- true
				reply := new(Reply)
				client.Go("Arith.Add", args, reply, res)
			}
		}()
		go func() {
			for call := range res {
				A := call.Args.(*Args).A
				B := call.Args.(*Args).B
				C := call.Reply.(*Reply).C
				if A+B != C {
					b.Errorf("incorrect reply: Add: expected %d got %d", A+B, C)
					return
				}
				<-gate
				if atomic.AddInt32(&recv, -1) == 0 {
					close(res)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
