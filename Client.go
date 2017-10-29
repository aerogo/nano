package nano

import (
	"log"
	"net"
)

// Client ...
type Client struct {
	connection *net.TCPConn
	close      chan bool
}

// connect ...
func (client *Client) connect() error {
	conn, err := net.Dial("tcp", "localhost:3000")

	if err != nil {
		return err
	}

	client.connection = conn.(*net.TCPConn)
	client.close = make(chan bool)

	go client.onConnectedToServer()

	return nil
}

// onConnectedToServer ...
func (client *Client) onConnectedToServer() {
	connection := client.connection

	for {
		select {
		case <-client.close:
			err := connection.Close()

			if err != nil {
				log.Fatal(err)
			}

			return
		}
	}

	// defer func() { client.closed = true }()
	// defer client.connection.Close()

	// for {
	// 	msg := make([]byte, 1024)
	// 	_, err := client.connection.Read(msg)

	// 	if err != nil {
	// 		continue
	// 	}

	// 	client.connection.Write([]byte("pong"))
	// }
}
