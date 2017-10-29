package nano

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/aerogo/packet"
)

// Client ...
type Client struct {
	packet.Stream
	close  chan bool
	closed bool
	db     *Database
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

	go client.read()
	go client.write()
	go client.readPackets()
	go client.waitClose()

	return nil
}

// read ...
func (client *Client) read() {
	err := client.Stream.Read()

	if err != nil {
		// fmt.Println(err)
	}
}

// write ...
func (client *Client) write() {
	err := client.Stream.Write()

	if err != nil {
		// fmt.Println(err)
	}
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

		case messageSet:
			onSet(msg, client.db)
		}
	}
}

// waitClose ...
func (client *Client) waitClose() {
	<-client.close

	client.closed = true
	err := client.Connection.Close()

	if err != nil {
		panic(err)
	}
}
