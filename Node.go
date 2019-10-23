package nano

import "net"

// Node is a general-purpose node in the cluster.
// It can act either as a server or as a client.
type Node interface {
	Address() net.Addr
	Close()
}

// node can dynamically switch between client and server mode.
type node struct {
	server
	client
}
