package nano

import (
	"net"
	"sync"

	"github.com/aerogo/nano/cluster"
	"github.com/aerogo/nano/packet"
)

// Node is a general-purpose node in the cluster.
type Node interface {
	Broadcast(msg packet.Packet) error
	Close()
	Namespace(string) Namespace
}

type node struct {
	*cluster.Node
	namespaces sync.Map
	config     Configuration
}

// New creates a new cluster node.
func New(config Configuration) Node {
	node := &node{config: config}
	node.Node = cluster.New(config.Port, node.onMessage)
	return node
}

// Namespace returns the namespace with the given name.
func (node *node) Namespace(name string) Namespace {
	obj, loaded := node.namespaces.Load(name)

	if !loaded {
		namespace, err := newNamespace(node, name)

		if err != nil {
			panic(err)
		}

		node.namespaces.Store(name, namespace)
		return namespace
	}

	return obj.(*namespace)
}

func (node *node) onMessage(address *net.UDPAddr, p packet.Packet) {
	switch p.Type() {
	case packetAlive:
		println("server is alive")

	default:
		println("unknown packet")
	}
}
