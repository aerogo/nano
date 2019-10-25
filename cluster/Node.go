package cluster

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aerogo/nano/packet"
)

// Node is a general-purpose node in the cluster.
type Node struct {
	reader    *net.UDPConn
	writer    *net.UDPConn
	address   *net.UDPAddr
	incoming  chan Message
	close     chan struct{}
	onMessage MessageHandler
}

// New creates a new cluster node.
func New(port int, messageHandler MessageHandler) *Node {
	if messageHandler == nil {
		messageHandler = func(*net.UDPAddr, packet.Packet) {}
	}

	node := &Node{
		close:     make(chan struct{}),
		incoming:  make(chan Message),
		onMessage: messageHandler,
	}

	err := node.init(port)

	if err != nil {
		panic(err)
	}

	return node
}

// init initializes the node.
func (node *Node) init(port int) error {
	var err error

	host := fmt.Sprintf("224.0.0.1:%d", port)
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

// read reads the packets and sends them to the "incoming" channel.
func (node *Node) read() {
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
			node.incoming <- Message{p, address}
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

// receive processes messages from the "incoming" channel.
func (node *Node) receive() {
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

// Broadcast sends a packet to the whole cluster.
func (node *Node) Broadcast(msg packet.Packet) error {
	_, err := node.writer.Write(msg)
	return err
}

// Close frees up all resources used by the node.
func (node *Node) Close() {
	close(node.close)
}
