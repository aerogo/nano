package nano

import (
	"net"
	"sync"
	"time"

	"github.com/aerogo/cluster"
	"github.com/aerogo/cluster/client"
	"github.com/aerogo/cluster/server"
	"github.com/aerogo/packet"
)

// Force interface implementation
var _ cluster.Node = (*Node)(nil)

// Node ...
type Node struct {
	namespaces  sync.Map
	node        cluster.Node
	server      *server.Node
	client      *client.Node
	port        int
	hosts       []string
	ioSleepTime time.Duration
}

// New ...
func New(port int, hosts ...string) *Node {
	// Create Node
	node := &Node{
		port:        port,
		hosts:       hosts,
		ioSleepTime: 1 * time.Millisecond,
	}

	node.connect()
	return node
}

// Namespace ...
func (node *Node) Namespace(name string, types ...interface{}) *Namespace {
	obj, found := node.namespaces.Load(name)

	if !found {
		namespace := NewNamespace(node, name, types...)
		node.namespaces.Store(name, namespace)
		return namespace
	}

	return obj.(*Namespace)
}

// IsServer ...
func (node *Node) IsServer() bool {
	return node.node.IsServer()
}

// IsClosed ...
func (node *Node) IsClosed() bool {
	return node.node.IsClosed()
}

// Broadcast ...
func (node *Node) Broadcast(msg *packet.Packet) {
	node.node.Broadcast(msg)
}

// Server ...
func (node *Node) Server() *server.Node {
	return node.server
}

// Client ...
func (node *Node) Client() *client.Node {
	return node.client
}

// Address ...
func (node *Node) Address() net.Addr {
	return node.node.Address()
}

// Clear deletes all data in the Node.
func (node *Node) Clear() {
	node.namespaces.Range(func(key, value interface{}) bool {
		namespace := value.(*Namespace)
		namespace.ClearAll()
		return true
	})
}

// Close ...
func (node *Node) Close() {
	node.node.Close()

	node.namespaces.Range(func(key, value interface{}) bool {
		namespace := value.(*Namespace)
		namespace.Close()
		return true
	})
}

// connect ...
func (node *Node) connect() {
	node.node = cluster.New(node.port, node.hosts...)

	if node.node.IsServer() {
		node.server = node.node.(*server.Node)
		node.server.OnConnect(serverOnConnect(node))
	} else {
		node.client = node.node.(*client.Node)
		go clientReadPackets(node.client, node)
	}
}

// broadcastRequired ...
func (node *Node) broadcastRequired() bool {
	if !node.IsServer() {
		return true
	}

	return node.server.ClientCount() > 0
}
