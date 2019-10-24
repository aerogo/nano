package nano

import (
	"fmt"
	"net"
	"time"

	"github.com/aerogo/nano/packet"
)

type client struct {
	connection *net.UDPConn
	incoming   chan packetWithAddress
	close      chan struct{}
}

type packetWithAddress struct {
	packet  packet.Packet
	address *net.UDPAddr
}

func (client *client) Address() net.Addr {
	return client.connection.LocalAddr()
}

func (client *client) Close() {
	close(client.close)
}

func (client *client) Main() {
	client.close = make(chan struct{})
	client.incoming = make(chan packetWithAddress)
	defer client.Close()

	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())

	go client.KeepAlive()
	go client.Receiver()

	buffer := make([]byte, 4096)

	for {
		n, address, err := client.connection.ReadFromUDP(buffer)
		p := packet.Packet(buffer[:n])
		client.incoming <- packetWithAddress{p, address}

		if err != nil {
			fmt.Printf("[%v] Error reading from UDP: %v\n", client.connection.LocalAddr(), err)
		}
	}
}

func (client *client) KeepAlive() {
	keepAlive := packet.New(packetAlive, nil)
	client.Send(keepAlive)

	ticker := time.NewTicker(keepAliveSendInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			client.Send(keepAlive)

		case <-client.close:
			return
		}
	}
}

func (client *client) Receiver() {
	for {
		select {
		case msg := <-client.incoming:
			fmt.Println("server is alive")
			client.Receive(msg.address, msg.packet)

		case <-time.After(keepAliveTimeout):
			fmt.Println("server is dead")

		case <-client.close:
			client.connection.Close()
			return
		}
	}
}

func (client *client) Send(p packet.Packet) {
	_, err := client.connection.Write(p)

	if err != nil {
		fmt.Printf("[%v] Error sending message to server: %v\n", client.connection.LocalAddr(), err)
	}
}

func (client *client) Receive(address *net.UDPAddr, p packet.Packet) {
	// fmt.Printf("[%v] Received message from %v of type %d: %s\n", client.connection.LocalAddr(), address, p.Type(), p.Data())

	switch p.Type() {
	case packetAlive:

	}
}
