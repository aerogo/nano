package nano

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

// Server ...
type Server struct {
	listener        *net.TCPListener
	connections     map[*net.TCPConn]*ServerConnection
	newConnections  chan *net.TCPConn
	deadConnections chan *net.TCPConn
	broadcasts      chan []byte
}

// start ...
func (server *Server) start() error {
	server.connections = make(map[*net.TCPConn]*ServerConnection)
	server.newConnections = make(chan *net.TCPConn, 32)
	server.deadConnections = make(chan *net.TCPConn, 32)
	server.broadcasts = make(chan []byte, 32)

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
			client := &ServerConnection{
				server:     server,
				connection: connection,
				incoming:   make(chan []byte),
				outgoing:   make(chan []byte),
			}

			server.connections[connection] = client

			go client.read()
			go client.write()

			fmt.Println("New connection", connection.RemoteAddr(), "#", len(server.connections))

			client.outgoing <- []byte("ping")

		case connection := <-server.deadConnections:
			close(server.connections[connection].incoming)
			close(server.connections[connection].outgoing)
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

// send ...
func (server *Server) send(conn *net.TCPConn, msg []byte) {
	totalWritten := 0

	for totalWritten < len(msg) {
		writtenThisCall, err := conn.Write(msg[totalWritten:])

		if err != nil {
			server.deadConnections <- conn
			break
		}

		totalWritten += writtenThisCall
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

// onConnection ...
func (server *Server) onConnection(connection *net.TCPConn) {
	for {
		_, err := connection.Write([]byte("ping"))

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("W Timeout", connection.RemoteAddr())
			break
		}

		if err != nil && err != io.EOF && strings.Contains(err.Error(), "connection reset") {
			fmt.Println("W Disconnected", connection.RemoteAddr())
			break
		}

		// connection.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		msg := make([]byte, 1024)
		_, err = connection.Read(msg)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("R Timeout", connection.RemoteAddr())
			break
		}

		if err != nil && err != io.EOF && strings.Contains(err.Error(), "connection reset") {
			fmt.Println("R Disconnected", connection.RemoteAddr())
			break
		}

		if err != nil {
			fmt.Println(string(msg))
		}
	}

	server.deadConnections <- connection
}
