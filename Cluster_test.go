package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

const nodeCount = 4

func TestClusterClose(t *testing.T) {
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)

		if i == 0 {
			assert.True(t, nodes[0].IsServer())
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].Close()
	}
}

func TestClusterDataSharing(t *testing.T) {
	// Create cluster where the server has initial data
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test", types...)

		if i == 0 {
			assert.True(t, nodes[0].IsServer())
			nodes[0].Namespace("test").Set("User", "100", newUser(100))
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	// Check data on client nodes
	for i := 1; i < nodeCount; i++ {
		user, err := nodes[i].Namespace("test").Get("User", "100")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func TestClusterSet(t *testing.T) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test", types...)

		if i == 0 {
			assert.True(t, nodes[i].IsServer())
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	// Make sure that node #0 does not have the record
	nodes[0].Namespace("test").Delete("User", "42")

	// Set record on node #1
	nodes[1].Namespace("test").Set("User", "42", newUser(42))

	// Wait until it propagates through the whole cluster
	time.Sleep(150 * time.Millisecond)

	// Confirm that all nodes have the record now
	for i := 0; i < nodeCount; i++ {
		user, err := nodes[i].Namespace("test").Get("User", "42")
		assert.NoError(t, err, "nodes[%d]", i)
		assert.NotNil(t, user, "nodes[%d]", i)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].Clear()
		nodes[i].Close()
	}
}
