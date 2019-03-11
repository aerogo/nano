package nano_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/nano"
)

func TestNodeAddress(t *testing.T) {
	node := nano.New(port)
	defer node.Close()
	assert.NotEmpty(t, node.Address().String())
}
