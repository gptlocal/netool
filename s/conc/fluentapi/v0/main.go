package main

import (
	"fmt"
	"github.com/sourcegraph/conc/iter"
)

func main() {
	input := []int{1, 2, 3, 4, 5}
	iterator := iter.Iterator[int]{MaxGoroutines: len(input) / 2}

	iterator.ForEach(input, func(v *int) {
		if *v%2 != 0 {
			*v = *v * 2
		}
	})

	fmt.Println(input)

	mapper := iter.Mapper[int, int]{MaxGoroutines: len(input) / 2}
	fmt.Println(mapper.Map(input, func(v *int) int {
		return *v * *v
	}))
}
