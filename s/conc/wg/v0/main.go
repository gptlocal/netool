package main

import (
	"fmt"
	"github.com/sourcegraph/conc"
	"sync"
)

func main() {
	origin()
	concv0()
}

func concv0() {
	wg := conc.NewWaitGroup()
	for i := 0; i < 10; i++ {
		wg.Go(doSomething)
	}
	wg.Wait()
}

func origin() {
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()

			// do something
			doSomething()
		}()
	}
	wg.Wait()
}

func doSomething() {
	fmt.Println("do something")
}
