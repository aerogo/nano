package nano

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aerogo/nano/packet"
	"github.com/akyoto/cache"
)

const (
	keepAliveSendInterval = 5 * time.Second
	keepAliveTimeout      = 2 * keepAliveSendInterval
	keepAliveCleanTime    = 1 * time.Minute
	readTimeout           = 5 * time.Minute
)

type server struct {
	listener *net.UDPConn
	clients  *cache.Cache
	incoming chan packetWithAddress
	close    chan struct{}
}

func (server *server) Address() net.Addr {
	return server.listener.LocalAddr()
}

func (server *server) Close() {
	close(server.close)
}

func (server *server) init(listener *net.UDPConn) {
	server.listener = listener
	server.close = make(chan struct{})
	server.incoming = make(chan packetWithAddress)
	server.clients = cache.New(keepAliveCleanTime)
	go server.Receiver()
	go server.Main()
}

func (server *server) Main() {
	defer fmt.Println("server.Main shutdown")
	defer close(server.incoming)

	buffer := make([]byte, 4096)
	fmt.Println("[server] Ready")

	for {
		server.listener.SetReadDeadline(time.Now().Add(readTimeout))
		n, address, err := server.listener.ReadFromUDP(buffer)

		if n > 0 {
			p := packet.Packet(buffer[:n])
			server.incoming <- packetWithAddress{p, address}
		}

		if err == nil {
			continue
		}

		netError, isNetError := err.(net.Error)

		if isNetError && netError.Timeout() {
			continue
		}

		// Go doesn't have a proper type for close errors,
		// so we need to do a string check which is not optimal.
		if !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Printf("[server] Error reading from UDP: %v\n", err)
		}

		return
	}
}

func (server *server) Receiver() {
	defer fmt.Println("server.Receiver shutdown")

	for {
		select {
		case msg := <-server.incoming:
			server.Receive(msg.address, msg.packet)

		case <-server.close:
			server.listener.Close()
			server.clients.Close()
			return
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
