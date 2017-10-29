package nano

import (
	"fmt"
	"log"
	"net"
)

// Client ...
type Client struct {
	PacketStream
	close chan bool
}

// connect ...
func (client *Client) connect() error {
	conn, err := net.Dial("tcp", "localhost:3000")

	if err != nil {
		return err
	}

	client.connection = conn.(*net.TCPConn)
	client.incoming = make(chan *Packet)
	client.outgoing = make(chan *Packet)
	client.close = make(chan bool)

	go client.read()
	go client.write()
	// go client.waitClose()

	return nil
}

// waitClose ...
func (client *Client) waitClose() {
	connection := client.connection

	// client.connection will be nil after we receive this.
	<-client.close
	fmt.Println("CLOSE")

	err := connection.Close()

	if err != nil {
		log.Fatal(err)
	}
}
