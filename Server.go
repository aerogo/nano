package nano

import (
	"fmt"
	"log"
	"net"
)

// Server ...
type Server struct {
	listener        *net.TCPListener
	connections     map[*net.TCPConn]*ServerConnection
	newConnections  chan *net.TCPConn
	deadConnections chan *net.TCPConn
	broadcasts      chan *Packet
}

// start ...
func (server *Server) start() error {
	server.connections = make(map[*net.TCPConn]*ServerConnection)
	server.newConnections = make(chan *net.TCPConn, 32)
	server.deadConnections = make(chan *net.TCPConn, 32)
	server.broadcasts = make(chan *Packet, 32)

	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		return err
	}

	server.listener = listener.(*net.TCPListener)

	go server.mainLoop()
	go server.acceptConnections()

	return nil
}

// mainLoop ...
func (server *Server) mainLoop() {
	for {
		select {
		case connection := <-server.newConnections:
			connection.SetNoDelay(true)

			client := &ServerConnection{
				server: server,
				PacketStream: PacketStream{
					connection: connection,
					incoming:   make(chan *Packet),
					outgoing:   make(chan *Packet),
				},
			}

			server.connections[connection] = client

			go client.read()
			go client.write()

			fmt.Println("New connection", connection.RemoteAddr(), "#", len(server.connections))

			// Send initial packet
			client.outgoing <- NewPacket(messageCollection, []byte("ping"))

		case connection := <-server.deadConnections:
			client, exists := server.connections[connection]

			if !exists {
				break
			}

			close(client.incoming)
			close(client.outgoing)
			connection.Close()
			delete(server.connections, connection)

		case msg := <-server.broadcasts:
			for connection := range server.connections {
				client := server.connections[connection]
				client.outgoing <- msg
			}
		}
	}
}

// acceptConnections ...
func (server *Server) acceptConnections() {
	fmt.Println("server", server.listener.Addr())

	for {
		conn, err := server.listener.Accept()

		if err != nil {
			log.Fatal("Accept error", err)
			continue
		}

		server.newConnections <- conn.(*net.TCPConn)
	}
}
