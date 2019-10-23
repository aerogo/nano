package nano

import (
	"fmt"
	"net"
)

// New creates a new nano node.
func New(config Configuration) Node {
	node := &node{}
	address := fmt.Sprintf(":%d", config.Port)
	udpAddress, err := net.ResolveUDPAddr("udp", address)

	if err != nil {
		panic(fmt.Errorf("Error resolving address: %v\n", err))
	}

	for {
		listener, err := net.ListenUDP("udp", udpAddress)

		if err != nil {
			connection, err := net.DialUDP("udp", nil, udpAddress)

			if err != nil {
				fmt.Printf("[client] Error connecting to server: %v\n", err)
				continue
			}

			node.client.connection = connection
			go node.client.Main()
			return node
		}

		node.listener = listener
		go node.server.Main()
		return node
	}
}
