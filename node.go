package nano

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aerogo/nano/packet"
)

// Node is a general-purpose node in the cluster.
// It can act either as a server or as a client.
type Node interface {
	Broadcast(msg packet.Packet) error
	Close()
}

// node can dynamically switch between client and server mode.
type node struct {
	reader   *net.UDPConn
	writer   *net.UDPConn
	address  *net.UDPAddr
	incoming chan packetWithAddress
	close    chan struct{}
	config   Configuration
}

// New creates a new nano node.
func New(config Configuration) Node {
	node := &node{
		config:   config,
		close:    make(chan struct{}),
		incoming: make(chan packetWithAddress),
	}

	err := node.init()

	if err != nil {
		panic(err)
	}

	return node
}

// init initializes the node.
func (node *node) init() error {
	var err error

	host := fmt.Sprintf("224.0.0.1:%d", node.config.Port)
	node.address, err = net.ResolveUDPAddr("udp", host)

	if err != nil {
		return err
	}

	node.reader, err = net.ListenMulticastUDP("udp", nil, node.address)

	if err != nil {
		return err
	}

	err = node.reader.SetReadBuffer(packet.BufferSize)

	if err != nil {
		return err
	}

	node.writer, err = net.DialUDP("udp", nil, node.address)

	if err != nil {
		return err
	}

	err = node.writer.SetWriteBuffer(packet.BufferSize)

	if err != nil {
		return err
	}

	go node.read()
	go node.receive()

	return nil
}

func (node *node) read() {
	defer fmt.Println("node.read shutdown")
	defer close(node.incoming)

	fmt.Printf("[%v] Successfully connected to %v\n", node.writer.LocalAddr(), node.writer.RemoteAddr())

	var (
		length  int
		address *net.UDPAddr
		buffer  = make([]byte, 4096)
	)

	for {
		err := node.reader.SetReadDeadline(time.Now().Add(readTimeout))

		if err != nil {
			goto errorHandler
		}

		length, address, err = node.reader.ReadFromUDP(buffer)

		// Skip messages from myself
		if address != nil && address.Port == node.address.Port && address.IP.Equal(node.address.IP) {
			continue
		}

		if length > 0 {
			p := packet.Packet(buffer[:length])
			node.incoming <- packetWithAddress{p, address}
		}

		if err == nil {
			continue
		}

	errorHandler:
		netError, isNetError := err.(net.Error)

		if isNetError && netError.Timeout() {
			continue
		}

		// Go doesn't have a proper type for close errors,
		// so we need to do a string check which is not optimal.
		if !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Printf("[%v] Error reading from UDP: %v\n", node.reader.LocalAddr(), err)
		}

		return
	}
}

func (node *node) receive() {
	defer fmt.Println("node.receiver shutdown")

	for {
		select {
		case msg := <-node.incoming:
			node.onMessage(msg.address, msg.packet)

		case <-node.close:
			node.reader.Close()
			return
		}
	}
}

func (node *node) Broadcast(msg packet.Packet) error {
	_, err := node.writer.Write(msg)
	return err
}

func (node *node) onMessage(address *net.UDPAddr, msg packet.Packet) {
	// fmt.Printf("[%v] Received message from %v of type %d: %s\n", node.reader.LocalAddr(), address, msg.Type(), msg.Data())

	switch msg.Type() {
	case packetAlive:
		fmt.Println("server is alive")
	}
}

// Close frees up all resources used by the node.
func (node *node) Close() {
	close(node.close)
}
