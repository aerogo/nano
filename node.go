package nano

import (
	"fmt"
	"net"
)

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
	config Configuration
}

// connect initializes the node as a server or client.
func (node *node) connect() error {
	address := fmt.Sprintf(":%d", node.config.Port)
	udpAddress, err := net.ResolveUDPAddr("udp", address)

	if err != nil {
		return err
	}

	listener, err := net.ListenUDP("udp", udpAddress)

	if err != nil {
		connection, err := net.DialUDP("udp", nil, udpAddress)

		if err != nil {
			return err
		}

		node.client.connection = connection
		go node.client.Main()
		return nil
	}

	node.server.listener = listener
	go node.server.Main()
	return nil
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
