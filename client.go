package nano

import (
	"fmt"
	"net"

	"github.com/aerogo/nano/packet"
)

type client struct {
	connection *net.UDPConn
}

func (client *client) Address() net.Addr {
	return client.connection.LocalAddr()
}

func (client *client) Close() {
	client.connection.Close()
}

func (client *client) Main() {
	defer client.Close()
	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())

	// for i := 0; i < 5; i++ {
	// 	client.Send(packet.New(packetAlive, nil))
	// }

	buffer := make([]byte, 4096)

	for {
		n, address, err := client.connection.ReadFromUDP(buffer)
		fmt.Printf("[%v] %s sent %d bytes\n", client.connection.LocalAddr(), address, n)

		p := packet.Packet(buffer[:n])
		client.Receive(address, p)

		if err != nil {
			fmt.Printf("[%v] Error reading from UDP: %v\n", client.connection.LocalAddr(), err)
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
	fmt.Printf("[%v] Received message from %v of type %d: %s\n", client.connection.LocalAddr(), address, p.Type(), p.Data())

	switch p.Type() {
	case packetAlive:
		fmt.Println(string(p.Data()))
	}
}
