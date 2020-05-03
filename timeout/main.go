package main

import (
	"fmt"
	"time"
)

// Does the same thing as time.After
func MyAfter(d time.Duration) chan time.Time {
	// needs to be buffered, so that the
	// timeout goroutint won't wait for the timout to be read
	// which will happen if a value is put on ch before the timeout
	// is reached.
	// This allows it to finish and be garbage collected.
	timeout := make(chan time.Time, 1)
	go func() {
		defer close(timeout)
		time.Sleep(d)
	}()
	return timeout
}

func main() {
	ch := make(chan string)
	go func() {
		// Don't timeout
		time.Sleep(1500 * time.Millisecond)
		ch <- "No timeout here!!!"
	}()
	select { // either read a value from ch, or timeout.
	case v := <-ch:
		fmt.Printf("Got a result! %s", v)
	case <-MyAfter(1 * time.Second):
		fmt.Println("Timed out... bye!")
		return

	}
}
