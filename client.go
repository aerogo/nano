package nano

import (
	"fmt"
	"net"
)

type client struct {
	connection *net.UDPConn
}

func (client *client) Main() {
	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())
	defer client.connection.Close()

	_, err := client.connection.Write([]byte("ping"))

	if err != nil {
		fmt.Printf("[%v] Error sending message to server: %v\n", client.connection.LocalAddr(), err)
	}

	_, err = client.connection.Write([]byte("ping"))

	if err != nil {
		fmt.Printf("[%v] Error sending message to server: %v\n", client.connection.LocalAddr(), err)
	}
}
