package nano

import (
	"fmt"
	"io"
	"net"
	"strings"
)

// ServerConnection ...
type ServerConnection struct {
	server     *Server
	incoming   chan []byte
	outgoing   chan []byte
	connection *net.TCPConn
}

// write ...
func (client *ServerConnection) write() {
	for msg := range client.outgoing {
		totalWritten := 0

		for totalWritten < len(msg) {
			writtenThisCall, err := client.connection.Write(msg[totalWritten:])

			if err != nil {
				client.server.deadConnections <- client.connection
				break
			}

			totalWritten += writtenThisCall
		}
	}
}

// read ...
func (client *ServerConnection) read() {
	for {
		msg := make([]byte, 1024)
		bytesRead, err := client.connection.Read(msg)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("R Timeout", client.connection.RemoteAddr())
			break
		}

		if err != nil && err != io.EOF && strings.Contains(err.Error(), "connection reset") {
			fmt.Println("R Disconnected", client.connection.RemoteAddr())
			break
		}

		client.incoming <- msg[:bytesRead]
	}

	client.server.deadConnections <- client.connection
}
