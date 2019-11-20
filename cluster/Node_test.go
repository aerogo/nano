package cluster_test

import (
	"net"
	"sync"
	"testing"

	"github.com/aerogo/nano/cluster"
	"github.com/aerogo/nano/packet"
	"github.com/akyoto/assert"
)

const (
	nodeCount = 2
	port      = 5000
)

func TestBroadcast(t *testing.T) {
	nodes := make([]*cluster.Node, nodeCount)
	waitGroup := sync.WaitGroup{}

	for i := range nodes {
		nodes[i] = cluster.New(port, func(*net.UDPAddr, packet.Packet) {
			waitGroup.Done()
		})

		waitGroup.Add(1)
	}

	err := nodes[0].Broadcast(packet.New(0, nil))
	assert.Nil(t, err)
	waitGroup.Wait()

	for _, node := range nodes {
		node.Close()
	}
}

func TestEmptyMessageHandler(t *testing.T) {
	cluster.New(port, nil).Close()
}
