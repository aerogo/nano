package nano

import (
	"fmt"
	"net"
)

// ServerConnection ...
type ServerConnection struct {
	PacketStream
	server *Server
}

// read ...
func (client *ServerConnection) read() {
	client.PacketStream.read()
	client.server.deadConnections <- client.connection.(*net.TCPConn)
}

// write ...
func (client *ServerConnection) write() {
	client.PacketStream.write()
	client.server.deadConnections <- client.connection.(*net.TCPConn)
}

// readPackets ...
func (client *ServerConnection) readPackets() {
	for packet := range client.incoming {
		switch packet.Type {
		case messagePong:
			fmt.Println(string(packet.Data))
		}
	}
}
