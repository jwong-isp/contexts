package main

import (
	"fmt"
	"sync"
	"time"
)

func wordGen(done <-chan struct{}, in <-chan int, prefix string) chan string {
	alphabet := "abcdefghij"
	out := make(chan string, 1)
	go func() {
		defer close(out)
		for v := range in {
			output := fmt.Sprintf("%s: %s", prefix, string(alphabet[v]))
			//fmt.Printf("wordGen output: %s\n", output)
			select {
			case out <- output:
			case <-done:
				fmt.Printf("wordGen, %s, kill received\n", prefix)
				return
			}
		}
		fmt.Printf("wordGen, %s, no more input from channel\n", prefix)
	}()
	return out
}

func numGen(done <-chan struct{}) chan int {
	out := make(chan int, 1)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			output := i % 10
			//fmt.Printf("numGen output: %s\n", output)
			select {
			case out <- output:
			case <-done:
				fmt.Println("numGen kill received")
				return
			}
		}
	}()
	return out
}

func merge(done <-chan struct{}, cs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan string) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				fmt.Println("Closing merge for output func")
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
		fmt.Println("Closed merge out")
	}()
	return out
}

func main() {
	done := make(chan struct{})
	nums := numGen(done)
	w1 := wordGen(done, nums, "w1")
	w2 := wordGen(done, nums, "w2")
	term := "w1: d"

	func() {
		defer close(done)
		for v := range merge(done, w1, w2) {
			if v == term {
				fmt.Printf("Found %s! Now the program can continue\n", term)
				/* Could also do this way, but that requires downstream receivers
				to know the number of upstream senders

				done <- struct{}{}
				done <- struct{}{}
				done <- struct{}{}
				*/
				/*
					Closing the channel means that all recievers of the channel will
					get a zero value when they try to consume.
				*/
				fmt.Println("Doing other stuff now, but all channels should be closed")
				return
			} else {
				fmt.Printf("'%s' is not '%s'\n", v, term)
			}
		}
	}()
	time.Sleep(3 * time.Second) // Wait for everything to close

}
