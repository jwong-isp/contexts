package main

import (
	"fmt"
	"sync"
	"time"
)

// wordGen reads integers from an input channel and sends a letter on an outbound channel.
// It ends when either the `done` or `in` channel is closed.
// It returns a channel that will receive letters prefixed with the `prefix` as strings.
func wordGen(done <-chan struct{}, in <-chan int, prefix string) chan string {
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	out := make(chan string, 1)
	go func() {
		defer close(out)
		for v := range in {
			output := fmt.Sprintf("%s: %s", prefix, string(alphabet[v]))
			select {
			case out <- output:
			case <-done:
				fmt.Printf("%s: Closing wordGen channel\n", prefix)
				return
			}
		}
		fmt.Printf("%s: wordGen no more input from numGen channel\n", prefix)
	}()
	return out
}

// numGen endlessly outputs integers between 0 and 10 on a channel
// It returns a channel that will recieve the generated numbers.
func numGen(done <-chan struct{}) chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			output := i % 26
			select {
			case out <- output:
			case <-done:
				fmt.Println("numGen: Closing output channel")
				return
			}
		}
	}()
	return out
}

// taken from https://blog.golang.org/pipelines
// Using this to show fan-in
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
				fmt.Println("merge: Stop consuming from input channel")
				// Also decrements the WaitGroup
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
		fmt.Println("merge: Closing the output channel")
	}()
	return out
}

func main() {
	done := make(chan struct{})
	nums := numGen(done)
	// fan-out the values for nums across multiple wordGens.
	w1 := wordGen(done, nums, "w1")
	w2 := wordGen(done, nums, "w2")
	term := "w1: z"

	/*
		Closing the channel means that all recievers of the channel will
		get a zero value when they try to consume.

		Here we will close `done` once the function exits. This could also be done
		in the `main()` function, but I wanted to show the channels getting the
		signal.
	*/
	// fan-in the wordGens using the `merge` function
	// Interesting note: pipelines are similar to python generators in this case
	mergeCh := merge(done, w1, w2)
	timeout := time.After(500 * time.Millisecond)
findLoop:
	for {
		select {
		case v := <-mergeCh:
			if v == term {
				fmt.Printf("main: Found %s! Continuing...\n", term)
				/* Could do this way instead of `defer close(done)`,
				but this requires downstream receivers to know the number of upstream senders

				done <- struct{}{}
				done <- struct{}{}
				done <- struct{}{}

				It is easier and more reliable to just close `done` once we are done consuming its output
				*/
				fmt.Println("main: Doing other stuff now, but all channels should be closed")
				close(done)
				break findLoop

			} else {
				fmt.Printf("main: '%s' is not '%s'\n", v, term)
				time.Sleep(100 * time.Millisecond)
			}
		case <-timeout:
			fmt.Printf("main: Timed out while searching for '%s' \n", term)
			close(done)
			break findLoop
		}

	}
	fmt.Println("main: Other stuff being done")
	time.Sleep(3 * time.Second) // Pretend we're doing other stuff for a while

}
