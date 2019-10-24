package nano

import (
	"fmt"
	"net"
	"strings"
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

func (client *client) init(connection *net.UDPConn) {
	client.connection = connection
	client.close = make(chan struct{})
	client.incoming = make(chan packetWithAddress)
	go client.KeepAlive()
	go client.Receiver()
	go client.Main()
}

func (client *client) Main() {
	defer fmt.Println("client.Main shutdown")
	defer close(client.incoming)

	fmt.Printf("[%v] Successfully connected to server %v\n", client.connection.LocalAddr(), client.connection.RemoteAddr())
	buffer := make([]byte, 4096)

	for {
		err := client.connection.SetReadDeadline(time.Now().Add(readTimeout))

		if err != nil {
			fmt.Printf("[%v] Error setting read deadline: %v\n", client.connection.LocalAddr(), err)
			return
		}

		n, address, err := client.connection.ReadFromUDP(buffer)

		if n > 0 {
			p := packet.Packet(buffer[:n])
			client.incoming <- packetWithAddress{p, address}
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
			fmt.Printf("[%v] Error reading from UDP: %v\n", client.connection.LocalAddr(), err)
		}

		return
	}
}

func (client *client) KeepAlive() {
	defer fmt.Println("client.KeepAlive shutdown")

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
	defer fmt.Println("client.Receiver shutdown")

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
