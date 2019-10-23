package nano

import (
	"fmt"
	"net"
)

type server struct {
	listener *net.UDPConn
	clients  map[string]*clientOnServer
}

type clientOnServer struct {
	buffer  []byte
	address *net.UDPAddr
}

func (server *server) Address() net.Addr {
	return server.listener.LocalAddr()
}

func (server *server) Close() {
	server.listener.Close()
}

func (server *server) Main() {
	server.clients = map[string]*clientOnServer{}
	buffer := make([]byte, 4096)

	for {
		n, address, err := server.listener.ReadFromUDP(buffer)
		client := server.clients[address.String()]

		if client == nil {
			client = &clientOnServer{
				address: address,
			}

			server.clients[address.String()] = client
		}

		client.buffer = append(client.buffer, buffer[:n]...)
		fmt.Printf("[server] %s buffer contains %d bytes (+%d)\n", client.address, len(client.buffer), n)

		if err != nil {
			fmt.Printf("[server] Error reading client packet: %v\n", err)
		}
	}
}
