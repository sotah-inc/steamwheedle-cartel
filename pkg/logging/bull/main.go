package main

import (
	"fmt"
)

func main() {
	in := make(chan int)
	go func() {
		for count := range in {
			fmt.Printf("count: %d\n", count)
		}
	}()

	for i := 0; i < 5; i += 1 {
		in <- i
	}

	fmt.Println("Hello, playground")
}
