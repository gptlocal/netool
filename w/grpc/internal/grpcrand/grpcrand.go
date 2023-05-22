package grpcrand

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r  = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu sync.Mutex
)

// Int implements rand.Int on the grpcrand global source.
func Int() int {
	mu.Lock()
	defer mu.Unlock()
	return r.Int()
}

// Int63n implements rand.Int63n on the grpcrand global source.
func Int63n(n int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	return r.Int63n(n)
}

// Intn implements rand.Intn on the grpcrand global source.
func Intn(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return r.Intn(n)
}

// Int31n implements rand.Int31n on the grpcrand global source.
func Int31n(n int32) int32 {
	mu.Lock()
	defer mu.Unlock()
	return r.Int31n(n)
}

// Float64 implements rand.Float64 on the grpcrand global source.
func Float64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return r.Float64()
}

// Uint64 implements rand.Uint64 on the grpcrand global source.
func Uint64() uint64 {
	mu.Lock()
	defer mu.Unlock()
	return r.Uint64()
}
