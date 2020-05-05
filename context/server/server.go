//TODO: Server handler to add request id
package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var id = 0

func nextRequestId() int {
	id++
	return id
}

func addRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "requestId", nextRequestId())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	requestId := r.Context().Value("requestId").(int)
	w.Header().Set("x-request-id", strconv.Itoa(requestId))
	fmt.Fprintf(w, "You called foo, request: %d", requestId)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("server: Whoops! This shouldn't have happened")
		fmt.Fprintln(w, "Whoops! This shouldn't have happened")
	case <-r.Context().Done():
		fmt.Println("server: Slow call canceled")
	}
}

func main() {
	http.Handle("/foo", addRequestId(http.HandlerFunc(fooHandler)))

	// Add a timeout to the request
	http.Handle("/slow", http.TimeoutHandler(http.HandlerFunc(slowHandler), 1*time.Second, "Too Slow! Bail!"))
	http.ListenAndServe(":8080", nil)
}
