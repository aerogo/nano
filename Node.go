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
	Clear()
	Close()
	Namespace(name string) Namespace
	NewNamespace(name string, types ...interface{}) Namespace
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

// Clear deletes the entire contents of the database.
func (node *node) Clear() {
	node.namespaces.Range(func(key, value interface{}) bool {
		namespace := value.(*namespace)
		namespace.ClearAll()
		return true
	})
}

// NewNamespace creates a new namespace.
func (node *node) NewNamespace(name string, types ...interface{}) Namespace {
	obj, exists := node.namespaces.Load(name)

	if exists {
		return obj.(Namespace)
	}

	namespace, err := newNamespace(node, name)

	if err != nil {
		panic(err)
	}

	namespace.RegisterTypes(types...)
	node.namespaces.Store(name, namespace)
	return namespace
}

// Namespace returns the namespace with the given name.
func (node *node) Namespace(name string) Namespace {
	obj, loaded := node.namespaces.Load(name)

	if !loaded {
		return nil
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
