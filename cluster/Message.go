package cluster

import (
	"net"

	"github.com/aerogo/nano/packet"
)

// Message represents a received message in the cluster.
type Message struct {
	packet  packet.Packet
	address *net.UDPAddr
}
