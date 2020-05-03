package main

import (
	"fmt"
	"sync"
	"time"
)

func blocking() {
	fmt.Println("Started")
	// this will block
	ch := make(chan int)
	defer close(ch)
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println(<-ch)
	}()
	ch <- 5
	fmt.Println("Finished after printing value")
}

func buffered() {
	// Use a WaitGroup so the function doesn't exit until
	// the goroutine finishes
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Println("Started")
	// this won't block
	ch := make(chan int, 1)
	defer close(ch)
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println(<-ch)
		wg.Done()
	}()
	ch <- 10 // Doesn't block
	fmt.Println("Finished before printing value.")
	wg.Wait()
}

func main() {
	fmt.Println("-----Non buffered-----")
	blocking()
	fmt.Println("\n\n-----Buffered-----")
	buffered()
}
