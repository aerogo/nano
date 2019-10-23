package nano

import (
	"fmt"
	"net"

	"github.com/aerogo/nano/packet"
)

type client struct {
	connection *net.UDPConn
	outgoing   chan packet.Packet
}

func (client *client) Address() net.Addr {
	return client.connection.LocalAddr()
}

func (client *client) Close() {
	client.connection.Close()
	close(client.outgoing)
}

func (client *client) Main() {
	client.outgoing = make(chan packet.Packet)
	defer client.Close()
	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())
	go client.Writer()
	client.outgoing <- packet.New(0, []byte("ping"))

	buffer := make([]byte, 4096)

	for {
		n, address, err := client.connection.ReadFromUDP(buffer)
		fmt.Printf("[%v] %s sent %d bytes\n", client.connection.LocalAddr(), address, n)

		p := packet.Packet(buffer[:n])
		client.OnPacket(address, p)

		if err != nil {
			fmt.Printf("[%v] Error reading from UDP: %v\n", client.connection.LocalAddr(), err)
		}
	}
}

func (client *client) Writer() {
	for packet := range client.outgoing {
		_, err := client.connection.Write(packet)

		if err != nil {
			fmt.Printf("[%v] Error sending message to server: %v\n", client.connection.LocalAddr(), err)
		}
	}
}

func (client *client) OnPacket(address *net.UDPAddr, p packet.Packet) {
	fmt.Printf("[%v] Received message from %v of type %d: %s\n", client.connection.LocalAddr(), address, p.Type(), p.Data())

	switch p.Type() {
	case 0:
		fmt.Println(string(p.Data()))
	}
}
