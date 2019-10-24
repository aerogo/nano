package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/aerogo/nano"
)

var nodes [2]nano.Node

func main() {
	numGoroutines := runtime.NumGoroutine()

	// Create nodes
	for i := range nodes {
		nodes[i] = nano.New(nano.Configuration{
			Port: 5000,
		})
	}

	// Wait for termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Close nodes
	for _, node := range nodes {
		node.Close()
	}

	runtime.GC()
	leakedGoroutines := runtime.NumGoroutine() - numGoroutines
	fmt.Printf("\n%d goroutines leaked\n", leakedGoroutines)
}
