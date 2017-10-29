package nano

// ServerConnection ...
type ServerConnection struct {
	PacketStream
	server *Server
}

// read ...
func (client *ServerConnection) read() {
	client.PacketStream.read()
	client.server.deadConnections <- client.connection
}

// write ...
func (client *ServerConnection) write() {
	client.PacketStream.write()
	client.server.deadConnections <- client.connection
}
