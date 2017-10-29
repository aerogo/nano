package nano

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/aerogo/packet"
)

// Server ...
type Server struct {
	db              *Database
	listener        *net.TCPListener
	connections     map[*net.TCPConn]*ServerConnection
	newConnections  chan *net.TCPConn
	deadConnections chan *net.TCPConn
	broadcasts      chan *packet.Packet
}

// start ...
func (server *Server) start() error {
	server.connections = make(map[*net.TCPConn]*ServerConnection)
	server.newConnections = make(chan *net.TCPConn, 32)
	server.deadConnections = make(chan *net.TCPConn, 32)
	server.broadcasts = make(chan *packet.Packet, 32)

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
			connection.SetKeepAlive(true)

			client := &ServerConnection{
				server: server,
				Stream: packet.Stream{
					Connection: connection,
					Incoming:   make(chan *packet.Packet),
					Outgoing:   make(chan *packet.Packet),
				},
			}

			server.connections[connection] = client

			go client.read()
			go client.write()
			go client.readPackets()

			fmt.Println("New connection", connection.RemoteAddr(), "#", len(server.connections))

			// Send initial packet
			client.Outgoing <- packet.New(messagePing, []byte("ping"))

			// Send collection data
			wg := sync.WaitGroup{}

			for typeName := range server.db.types {
				wg.Add(1)

				go func(name string) {
					collection := server.db.Collection(name)

					var b bytes.Buffer
					b.WriteString(collection.name)
					b.WriteByte('\n')

					writer := bufio.NewWriter(&b)
					collection.writeRecords(writer, false)
					writer.Flush()

					client.Outgoing <- packet.New(messageCollection, b.Bytes())

					wg.Done()
				}(typeName)
			}

			wg.Wait()

		case connection := <-server.deadConnections:
			client, exists := server.connections[connection]

			if !exists {
				break
			}

			close(client.Incoming)
			close(client.Outgoing)
			connection.Close()
			delete(server.connections, connection)

		case msg := <-server.broadcasts:
			for _, client := range server.connections {
				client.Outgoing <- msg
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
			continue
		}

		server.newConnections <- conn.(*net.TCPConn)
	}
}
