package nano

import (
	"fmt"
	"net"

	"github.com/aerogo/packet"
)

// ServerConnection ...
type ServerConnection struct {
	packet.Stream
	server *Server
}

// read ...
func (client *ServerConnection) read() {
	client.Stream.Read()
	client.server.deadConnections <- client.Connection.(*net.TCPConn)
}

// write ...
func (client *ServerConnection) write() {
	client.Stream.Write()
	client.server.deadConnections <- client.Connection.(*net.TCPConn)
}

// readPackets ...
func (client *ServerConnection) readPackets() {
	for msg := range client.Incoming {
		switch msg.Type {
		case messagePong:
			fmt.Println(string(msg.Data))
		}
	}
}
