package test

import (
	"fmt"
	"runtime"
	"time"
)

// StackPrinter helps to debug tests which are hanging by printing stacks at a provided interval.
func StackPrinter(interval time.Duration) {
	buff := make([]byte, 10000)

	ticker := time.NewTicker(interval)
	for range ticker.C {
		fmt.Printf("%v\n", string(buff[:runtime.Stack(buff, true)]))
	}
}
