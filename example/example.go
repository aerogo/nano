package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aerogo/nano"
)

var nodes [3]nano.Node

func main() {
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
}
