package nano

import (
	"fmt"
	"net"
	"os/user"
	"path"
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
	config             Configuration
	ioSleepTime        time.Duration
	networkWorkerQueue chan *packet.Packet
	verbose            bool
}

// New starts up a new database node.
func New(config Configuration) *Node {
	// Get user info to access the home directory
	user, err := user.Current()

	if err != nil {
		panic(err)
	}

	// Create Node
	node := &Node{
		config:             config,
		ioSleepTime:        1 * time.Millisecond,
		networkWorkerQueue: make(chan *packet.Packet, 8192),
	}

	if node.config.Directory == "" {
		node.config.Directory = path.Join(user.HomeDir, ".aero", "db")
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

// Close frees up resources used by the node.
func (node *Node) Close() {
	if node.IsServer() {
		if node.verbose {
			fmt.Println("[server] broadcast close")
		}

		node.Broadcast(packet.New(packetServerClose, nil))
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
	node.node = cluster.New(node.config.Port, node.config.Hosts...)

	if node.node.IsServer() {
		node.server = node.node.(*server.Node)
		node.server.OnConnect(serverOnConnect(node))
	} else {
		node.client = node.node.(*client.Node)
		go clientReadPacketsFromServer(node.client, node)

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
