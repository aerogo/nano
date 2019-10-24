package nano

import (
	"fmt"
	"net"
	"time"

	"github.com/aerogo/nano/packet"
	"github.com/akyoto/cache"
)

const (
	keepAliveSendInterval = 5 * time.Second
	keepAliveTimeout      = 2 * keepAliveSendInterval
	keepAliveCleanTime    = 1 * time.Minute
)

type server struct {
	listener *net.UDPConn
	clients  *cache.Cache
}

func (server *server) Address() net.Addr {
	return server.listener.LocalAddr()
}

func (server *server) Close() {
	server.listener.Close()
}

func (server *server) Main() {
	server.clients = cache.New(keepAliveCleanTime)
	defer server.Close()
	buffer := make([]byte, 4096)
	fmt.Println("[server] Ready to receive messages")

	for {
		n, address, err := server.listener.ReadFromUDP(buffer)

		p := packet.Packet(buffer[:n])
		server.Receive(address, p)

		if err != nil {
			fmt.Printf("[server] Error reading from UDP: %v\n", err)
		}
	}
}

// Sends sends the packet to the given address.
func (server *server) Send(address *net.UDPAddr, p packet.Packet) {
	_, err := server.listener.WriteToUDP(p, address)

	if err != nil {
		fmt.Printf("[server] Error writing to %s: %v\n", address, err)
	}
}

// Broadcast sends a packet to all clients.
func (server *server) Broadcast(p packet.Packet) error {
	var err error

	server.clients.Range(func(key interface{}, value interface{}) bool {
		address := value.(*net.UDPAddr)
		_, err = server.listener.WriteToUDP(p, address)
		return err == nil
	})

	return err
}

// Receive handles received packets.
func (server *server) Receive(address *net.UDPAddr, msg packet.Packet) {
	// Refresh keep-alive timer
	server.clients.Set(address.String(), address, keepAliveTimeout)

	switch msg.Type() {
	case packetAlive:
		server.Send(address, packet.New(packetAlive, nil))
	}
}
