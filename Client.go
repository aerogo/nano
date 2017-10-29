package nano

import (
	"fmt"
	"log"
	"net"

	"github.com/aerogo/packet"
)

// Client ...
type Client struct {
	packet.Stream
	close chan bool
}

// connect ...
func (client *Client) connect() error {
	conn, err := net.Dial("tcp", "localhost:3000")

	if err != nil {
		return err
	}

	client.Connection = conn.(*net.TCPConn)
	client.Incoming = make(chan *packet.Packet)
	client.Outgoing = make(chan *packet.Packet)
	client.close = make(chan bool)

	go client.Read()
	go client.Write()
	go client.readPackets()
	go client.waitClose()

	return nil
}

// readPackets ...
func (client *Client) readPackets() {
	for msg := range client.Incoming {
		switch msg.Type {
		case messagePing:
			fmt.Println(string(msg.Data))
			client.Outgoing <- packet.New(messagePong, []byte("pong"))
		}
	}
}

// waitClose ...
func (client *Client) waitClose() {
	connection := client.Connection

	// client.connection will be nil after we receive this.
	<-client.close

	err := connection.Close()

	if err != nil {
		log.Fatal(err)
	}
}
