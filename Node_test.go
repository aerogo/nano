package nano_test

import (
	"testing"

	"github.com/aerogo/nano"
	"github.com/akyoto/assert"
)

func TestNodeAddress(t *testing.T) {
	node := nano.New(config)
	defer node.Close()
	assert.True(t, node.Address().String() != "")
}
