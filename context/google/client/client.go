package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	// Add a timeout to the context. No request should take longer than 2 seconds
	newCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a new request that will take 5 seconds to complete.
	req, _ := http.NewRequestWithContext(
		newCtx, http.MethodGet, "https://httpstat.us/418?sleep=5000", nil)

	// Make the request. This will return with an error because the context
	// deadline of 2 seconds is shorter than the request duration of 5 seconds.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		return
	}
	defer res.Body.Close()
	text, _ := ioutil.ReadAll(res.Body)

	fmt.Printf("response: %s", text)
}
