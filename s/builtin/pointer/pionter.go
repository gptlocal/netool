package main

import "fmt"

type MyStruct struct {
	i int
	s string
}

func main() {
	p1 := &MyStruct{1, "hello"}
	var p2 *MyStruct

	p2 = &MyStruct{2, "world"}

	fmt.Printf("p1: %p, p2: %p\n", p1, p2)
	fmt.Printf("p1: %v, p2: %v\n", p1, p2)

	*p1 = MyStruct{}
	fmt.Printf("p1: %p, p2: %p\n", p1, p2)
	fmt.Printf("p1: %v, p2: %v\n", p1, p2)

	*p1 = *p2
	fmt.Printf("p1: %p, p2: %p\n", p1, p2)
	fmt.Printf("p1: %v, p2: %v\n", p1, p2)

	p1 = p2
	fmt.Printf("p1: %p, p2: %p\n", p1, p2)
	fmt.Printf("p1: %v, p2: %v\n", p1, p2)
}
