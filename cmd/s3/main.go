package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	parent := context.Background()

	ctx1, cancel1 := context.WithTimeout(parent, 1*time.Second)
	defer cancel1()

	ctx2, cancel2 := context.WithTimeout(ctx1, 5*time.Second)
	defer cancel2()

	select {
	case <-time.After(3 * time.Second):
		fmt.Println("overslept")
	case <-ctx1.Done():
		fmt.Printf("ctx1: %s", ctx1.Err()) // prints "context deadline exceeded"
	case <-ctx2.Done():
		fmt.Printf("ctx2: %s", ctx2.Err()) // prints "context deadline exceeded"
	}

	return
}
