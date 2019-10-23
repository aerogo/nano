package nano

import (
	"fmt"
	"net"

	"github.com/aerogo/nano/packet"
)

type client struct {
	connection *net.UDPConn
}

func (client *client) Main() {
	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())
	defer client.connection.Close()

	_, err := client.connection.Write(packet.New(0, []byte("ping")))

	if err != nil {
		fmt.Printf("[%v] Error sending message to server: %v\n", client.connection.LocalAddr(), err)
	}

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

func (client *client) OnPacket(address *net.UDPAddr, p packet.Packet) {
	fmt.Printf("[%v] Received message from %v of type %d: %s\n", client.connection.LocalAddr(), address, p.Type(), p.Data())

	switch p.Type() {
	case 0:
		fmt.Println(string(p.Data()))
	}
}
