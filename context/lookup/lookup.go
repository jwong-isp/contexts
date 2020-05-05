package main

import (
	"context"
	"fmt"
)

func main() {
	parent := context.WithValue(context.Background(), "a", 1)
	child := context.WithValue(parent, "b", 2)
	grandchild := context.WithValue(child, "c", 3)
	grandchild2, cancel := context.WithCancel(child)
	defer cancel()

	fmt.Printf("%s: %d\n", "grandchild's a", grandchild.Value("a"))
	fmt.Printf("%s: %d\n", "grandchild's b", grandchild.Value("b"))
	fmt.Printf("%s: %d\n", "grandchild's c", grandchild.Value("c"))
	fmt.Printf("%s: %d\n", "grandchild2's c", grandchild2.Value("c"))
	fmt.Printf("%s: %d\n", "grandchild2's b", grandchild2.Value("b"))

}
