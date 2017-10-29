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
	go client.readPackets()
	go client.waitClose()

	return nil
}

// readPackets ...
func (client *Client) readPackets() {
	for packet := range client.incoming {
		switch packet.Type {
		case messagePing:
			fmt.Println(string(packet.Data))
			client.outgoing <- NewPacket(messagePong, []byte("pong"))
		}
	}
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
