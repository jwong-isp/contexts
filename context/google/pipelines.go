package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// wordGen reads integers from an input channel and sends a letter on an outbound channel.
// It ends when either the `done` or `in` channel is closed.
// It returns a channel that will receive letters prefixed with the `prefix` as strings.
func wordGen(ctx context.Context, in <-chan int) chan string {
	prefix := ctx.Value("prefix").(string)
	alphabet := "abcdefghijklmnopqrstuvwxyz"
	out := make(chan string, 1)
	go func() {
		defer close(out)
		for v := range in {
			output := fmt.Sprintf("%s: %s", prefix, string(alphabet[v]))
			select {
			case out <- output:
			case <-ctx.Done():
				fmt.Printf("%s: Closing wordGen channel\n", prefix)
				return
			}
		}
		fmt.Printf("%s: wordGen no more input from numGen channel\n", prefix)
	}()
	return out
}

// numGen endlessly outputs integers between 0 and 26 on a channel
// It returns a channel that will recieve the generated numbers.
func numGen(ctx context.Context) chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; ; i++ {
			output := i % 26
			//fmt.Printf("numGen output: %s\n", output)
			select {
			case out <- output:
			case <-ctx.Done():
				fmt.Println("numGen: Closing output channel")
				return
			}
		}
	}()
	return out
}

// taken from https://blog.golang.org/pipelines
// Using this to show fan-in
func merge(ctx context.Context, cs ...<-chan string) <-chan string {
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
			case <-ctx.Done():
				fmt.Println("merge: Stop consuming from fan-in input channel")
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
		fmt.Println("merge: Closing the fan-in output channel")
	}()
	return out
}

func main() {
	// Don't search longer than 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctxW1Val := context.WithValue(ctx, "prefix", "w1")
	ctxW2Val := context.WithValue(ctx, "prefix", "w2")
	nums := numGen(ctx)
	// fan-out the values for nums across multiple wordGens.
	ctxW1ValTimeout, _ := context.WithTimeout(ctxW1Val, 10000*time.Second)
	w1 := wordGen(ctxW1ValTimeout, nums) // prefix now comes from context

	ctxW2Timeout, _ := context.WithTimeout(ctxW2Val, 1*time.Second)
	w2 := wordGen(ctxW2Timeout, nums)

	term := "w1: z"
	// fan-in the wordGens using the `merge` function
	for v := range merge(ctx, w1, w2) {
		if v == term {
			fmt.Printf("main: Found %s! Continuing...\n", term)
			fmt.Println("Doing other stuff now, but all channels should be closed")
			// Cancel the context instead of closing a channel
			cancel()
			break

		} else {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("'%s' is not '%s'\n", v, term)
		}
	}
	fmt.Println("Finished doing other stuff!")
	time.Sleep(5 * time.Second) // Pretend we're doing other stuff for a while

}
