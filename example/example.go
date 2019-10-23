package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aerogo/nano"
)

const (
	nodeCount = 2
)

func main() {
	for i := 0; i < nodeCount; i++ {
		nano.New(nano.Configuration{
			Port: 5000,
		})
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
