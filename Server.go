package nano

import (
	"fmt"
	"io"
	"net"
	"strings"
)

// Server ...
type Server struct {
	listener    *net.TCPListener
	connections []*net.TCPConn
}

// start ...
func (server *Server) start() error {
	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		return err
	}

	server.listener = listener.(*net.TCPListener)
	go server.acceptConnections()

	return nil
}

// acceptConnections ...
func (server *Server) acceptConnections() {
	fmt.Println("server", server.listener.Addr())

	for {
		conn, err := server.listener.Accept()

		if err != nil {
			fmt.Println("server: Connection error", err)
			continue
		}

		go server.onConnection(conn.(*net.TCPConn))
	}
}

// onConnection ...
func (server *Server) onConnection(connection *net.TCPConn) {
	defer connection.Close()

	connection.SetNoDelay(true)
	server.connections = append(server.connections, connection)

	fmt.Println("New connection", connection.RemoteAddr(), "#", len(server.connections))

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
}
