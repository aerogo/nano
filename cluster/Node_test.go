package cluster_test

import (
	"runtime"
	"testing"

	"github.com/aerogo/nano/cluster"
	"github.com/akyoto/assert"
)

var (
	nodes [2]*cluster.Node
	port  = 5000
)

func TestGoroutineLeak(t *testing.T) {
	numGoroutines := runtime.NumGoroutine()

	for i := range nodes {
		nodes[i] = cluster.New(port, nil)
	}

	for _, node := range nodes {
		node.Close()
	}

	runtime.Gosched()
	runtime.GC()

	leakedGoroutines := runtime.NumGoroutine() - numGoroutines
	assert.Equal(t, leakedGoroutines, 0)
}
