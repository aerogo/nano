package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/flow"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

const nodeCount = 4
const parallelRequestCount = 8

func TestClusterClose(t *testing.T) {
	nodes := make([]*nano.Node, nodeCount)

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

	// Close clients
	for i := nodeCount - 1; i >= 0; i-- {
		assert.False(t, nodes[i].IsClosed())
		nodes[i].Close()
		assert.True(t, nodes[i].IsClosed())
	}
}

func TestClusterReconnect(t *testing.T) {
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)

		if i == 0 {
			assert.True(t, nodes[0].IsServer())
			nodes[0].Namespace("test").Set("User", "1", newUser(1))
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	// Wait for clients to connect
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)
	}

	nodes[2].Namespace("test").Set("User", "2", newUser(2))
	assert.True(t, nodes[2].Namespace("test").Exists("User", "2"))

	// Close server only
	nodes[0].Close()

	// Wait a bit, to test some real downtime
	time.Sleep(1500 * time.Millisecond)

	// Restart server
	nodes[0] = nano.New(port)
	nodes[0].Namespace("test").RegisterTypes(types...)
	assert.True(t, nodes[0].IsServer())

	// Wait for clients to reconnect in the span of a few seconds
	start := time.Now()
	for nodes[0].Server().ClientCount() < nodeCount-1 {
		time.Sleep(10 * time.Millisecond)

		if time.Since(start) > 2*time.Second {
			assert.Fail(t, "Not enough clients reconnected")
			return
		}
	}

	for i := 0; i < nodeCount; i++ {
		obj, err := nodes[i].Namespace("test").Get("User", "1")
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, "1", obj.(*User).ID)

		obj, err = nodes[i].Namespace("test").Get("User", "2")
		assert.NoError(t, err)
		assert.NotNil(t, obj)
		assert.Equal(t, "2", obj.(*User).ID)
	}

	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Close()
	}
}

func TestClusterDataSharing(t *testing.T) {
	// Create cluster where the server has initial data
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)

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
		flow.ParallelRepeat(parallelRequestCount, func() {
			user, err := nodes[i].Namespace("test").Get("User", "100")
			assert.NoError(t, err)
			assert.NotNil(t, user)
		})
	}

	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func TestClusterSet(t *testing.T) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)

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
	flow.ParallelRepeat(parallelRequestCount, func() {
		nodes[0].Namespace("test").Delete("User", "42")
	})

	// Set record on node #1
	flow.ParallelRepeat(parallelRequestCount, func() {
		nodes[1].Namespace("test").Set("User", "42", newUser(42))
	})

	// Wait until it propagates through the whole cluster
	time.Sleep(300 * time.Millisecond)

	// Confirm that all nodes have the record now
	for i := 0; i < nodeCount; i++ {
		user, err := nodes[i].Namespace("test").Get("User", "42")
		assert.NoError(t, err, "nodes[%d]", i)
		assert.NotNil(t, user, "nodes[%d]", i)
	}

	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func TestClusterDelete(t *testing.T) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New(port)
		nodes[i].Namespace("test").RegisterTypes(types...)

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

	// Delete on all nodes
	flow.ParallelRepeat(parallelRequestCount, func() {
		nodes[2].Namespace("test").Delete("User", "42")
	})

	// Wait until it propagates through the whole cluster
	time.Sleep(150 * time.Millisecond)

	// Confirm that all nodes deleted the record now
	for i := 0; i < nodeCount; i++ {
		exists := nodes[i].Namespace("test").Exists("User", "42")
		assert.False(t, exists, "nodes[%d]", i)
	}

	for i := nodeCount - 1; i >= 0; i-- {
		nodes[i].Clear()
		nodes[i].Close()
	}
}
