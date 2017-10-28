package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/nano"
)

func TestClientClose(t *testing.T) {
	nodeCount := 5
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)
	}

	time.Sleep(100 * time.Millisecond)

	for i := 1; i < nodeCount; i++ {
		nodes[i].Close()
	}

	time.Sleep(100 * time.Millisecond)
	nodes[0].Close()
}
