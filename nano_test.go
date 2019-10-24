package nano_test

import (
	"runtime"
	"testing"

	"github.com/aerogo/nano"
	"github.com/akyoto/assert"
)

var (
	nodes  [2]nano.Node
	config = nano.Configuration{
		Port: 5000,
	}
)

func TestGoroutineLeak(t *testing.T) {
	numGoroutines := runtime.NumGoroutine()

	for i := range nodes {
		nodes[i] = nano.New(config)
	}

	for _, node := range nodes {
		node.Close()
	}

	runtime.Gosched()
	runtime.GC()
	leakedGoroutines := runtime.NumGoroutine() - numGoroutines
	assert.Equal(t, leakedGoroutines, 0)
}

func TestNodeAddress(t *testing.T) {
	for i := range nodes {
		nodes[i] = nano.New(config)
		assert.NotNil(t, nodes[i].Address())
	}

	for _, node := range nodes {
		node.Close()
	}
}
