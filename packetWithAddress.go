package nano

import (
	"net"

	"github.com/aerogo/nano/packet"
)

type packetWithAddress struct {
	packet  packet.Packet
	address *net.UDPAddr
}
