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

func (node *node) Address() net.Addr {
	if node.listener != nil {
		return node.server.Address()
	} else {
		return node.client.Address()
	}
}

func (node *node) Close() {
	if node.listener != nil {
		node.server.Close()
	} else {
		node.client.Close()
	}
}
