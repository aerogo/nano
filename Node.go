package nano

import (
	"net"

	"github.com/aerogo/nano/cluster"
	"github.com/aerogo/nano/packet"
)

// Node is a general-purpose node in the cluster.
type Node interface {
	Broadcast(msg packet.Packet) error
	Close()
	Namespace(string)
}

type node struct {
	*cluster.Node
	config Configuration
}

// New creates a new cluster node.
func New(config Configuration) Node {
	node := &node{config: config}
	node.Node = cluster.New(config.Port, node.onMessage)
	return node
}

// Namespace
func (node *node) Namespace(name string) {
}

func (node *node) onMessage(address *net.UDPAddr, p packet.Packet) {
	switch p.Type() {
	case packetAlive:
		println("server is alive")
	}
}
