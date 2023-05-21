package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	limiter := rate.NewLimiter(100, 1)
	var i, j int32

	c := time.After(10 * time.Second)
	var wg sync.WaitGroup
LOOP:
	for {
		select {
		case <-c:
			break LOOP
		default:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if limiter.Allow() {
					atomic.AddInt32(&i, 1)
				} else {
					atomic.AddInt32(&j, 1)
				}
			}()
		}

		//if atomic.LoadInt32(&j)%100 == 0 {
		//	time.Sleep(time.Millisecond)
		//}
	}

	wg.Wait()
	fmt.Println("i:", i, "j:", j)
}
