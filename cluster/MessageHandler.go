package cluster

import (
	"net"

	"github.com/aerogo/nano/packet"
)

// MessageHandler is the function signature for functions that receive messages.
type MessageHandler func(*net.UDPAddr, packet.Packet)
