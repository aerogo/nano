package nano

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/aerogo/cluster"
	"github.com/aerogo/cluster/client"
	"github.com/aerogo/cluster/server"
	"github.com/aerogo/packet"
)

// Force interface implementation
var _ cluster.Node = (*Node)(nil)

// Node represents a single database node in the cluster.
type Node struct {
	namespaces         sync.Map
	node               cluster.Node
	server             *server.Node
	client             *client.Node
	port               int
	hosts              []string
	ioSleepTime        time.Duration
	networkSetQueue    chan *packet.Packet
	networkDeleteQueue chan *packet.Packet
	verbose            bool
}

// New ...
func New(port int, hosts ...string) *Node {
	// Create Node
	node := &Node{
		port:               port,
		hosts:              hosts,
		ioSleepTime:        1 * time.Millisecond,
		networkSetQueue:    make(chan *packet.Packet, 8192),
		networkDeleteQueue: make(chan *packet.Packet, 8192),
	}

	node.connect()
	return node
}

// Namespace ...
func (node *Node) Namespace(name string) *Namespace {
	obj, loaded := node.namespaces.LoadOrStore(name, nil)

	if !loaded {
		namespace := newNamespace(node, name)
		node.namespaces.Store(name, namespace)
		return namespace
	}

	// Wait for existing namespace load
	for obj == nil {
		time.Sleep(1 * time.Millisecond)
		obj, _ = node.namespaces.Load(name)
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
	if node.IsServer() {
		if node.verbose {
			fmt.Println("[server] broadcast close")
		}

		node.Broadcast(packet.New(packetClose, nil))
	}

	// Close cluster node
	node.node.Close()

	// Close namespaces
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

		for i := 0; i < runtime.NumCPU(); i++ {
			go clientNetworkWorker(node)
		}
	}
}

// broadcastRequired ...
func (node *Node) broadcastRequired() bool {
	if !node.IsServer() {
		return true
	}

	return node.server.ClientCount() > 0
}
