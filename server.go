package nano

import (
	"fmt"
	"net"

	"github.com/aerogo/nano/packet"
)

type server struct {
	listener *net.UDPConn
}

func (server *server) Address() net.Addr {
	return server.listener.LocalAddr()
}

func (server *server) Close() {
	server.listener.Close()
}

func (server *server) Main() {
	defer server.Close()
	buffer := make([]byte, 4096)

	for {
		n, address, err := server.listener.ReadFromUDP(buffer)
		fmt.Printf("[server] %s sent %d bytes\n", address, n)

		p := packet.Packet(buffer[:n])
		server.OnPacket(address, p)

		if err != nil {
			fmt.Printf("[server] Error reading from UDP: %v\n", err)
		}
	}
}

func (server *server) OnPacket(address *net.UDPAddr, p packet.Packet) {
	fmt.Printf("[server] %s message of type %d: %s\n", address, p.Type(), p.Data())

	switch p.Type() {
	case 0:
		_, err := server.listener.WriteToUDP(packet.New(0, []byte("pong")), address)

		if err != nil {
			fmt.Printf("[server] Error writing to %s: %v\n", address, err)
		}
	}
}
