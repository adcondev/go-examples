package main

import (
	"fmt"
	"time"
)

func panicCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic recovered! %v\n", r)
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("Panic Cleanup!\n")
			panic(r)
		}
	}()
	panic("PANIC!\n")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Graceful shutdown!")
		}
	}()
	fmt.Printf("Panic incoming!\n")
	panicCall()
}
