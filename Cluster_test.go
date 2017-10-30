package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

const nodeCount = 5

func TestClusterDataSharing(t *testing.T) {
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)

		if i == 0 {
			assert.True(t, nodes[i].Node().IsServer())
			nodes[i].Set("User", "100", newUser(100))
		} else {
			assert.False(t, nodes[i].Node().IsServer())
		}
	}

	time.Sleep(200 * time.Millisecond)

	for i := 1; i < nodeCount; i++ {
		user, err := nodes[i].Get("User", "100")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].ClearAll()
		nodes[i].Close()
	}
}

func TestClusterBroadcast(t *testing.T) {
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)

		if i == 0 {
			assert.True(t, nodes[i].Node().IsServer())
		} else {
			assert.False(t, nodes[i].Node().IsServer())
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Make sure that node #0 does not have the record
	nodes[0].Delete("User", "42")

	// Set record on node #2
	nodes[2].Set("User", "42", newUser(42))
	time.Sleep(200 * time.Millisecond)

	// Confirm that all nodes have the record now
	for i := 0; i < nodeCount; i++ {
		user, err := nodes[i].Get("User", "42")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].ClearAll()
		nodes[i].Close()
	}
}
