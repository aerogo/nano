package nano

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aerogo/packet"
)

// Client ...
type Client struct {
	packet.Stream
	close chan bool
	db    *Database
}

// connect ...
func (client *Client) connect() error {
	conn, err := net.Dial("tcp", "localhost:3000")

	if err != nil {
		return err
	}

	client.Connection = conn
	client.Incoming = make(chan *packet.Packet)
	client.Outgoing = make(chan *packet.Packet)
	client.close = make(chan bool)

	conn.(*net.TCPConn).SetNoDelay(true)
	conn.(*net.TCPConn).SetKeepAlive(true)

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

		case messageCollection:
			data := bytes.NewBuffer(msg.Data)
			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := client.db.Collection(collectionName)
			collection.readRecords(data)
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
