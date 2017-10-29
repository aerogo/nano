package nano

import (
	"log"
	"net"
)

// Client ...
type Client struct {
	connection *net.TCPConn
	incoming   chan []byte
	outgoing   chan []byte
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
	client.incoming = make(chan []byte)
	client.outgoing = make(chan []byte)

	go client.waitClose()

	return nil
}

// waitClose ...
func (client *Client) waitClose() {
	connection := client.connection

	// client.connection will be nil after we receive this.
	<-client.close

	err := connection.Close()

	if err != nil {
		log.Fatal(err)
	}
}
