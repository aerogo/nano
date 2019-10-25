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
}

// New creates a new cluster node.
func New(config Configuration) Node {
	return &node{
		Node: cluster.New(config.Port, onMessage),
	}
}

// Namespace
func (node *node) Namespace(name string) {
}

func onMessage(address *net.UDPAddr, p packet.Packet) {
	switch p.Type() {
	case packetAlive:
		println("server is alive")
	}
}
