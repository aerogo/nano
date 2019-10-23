package packet

// Packet is a byte slice abstraction to send types of messages
// within a packet.
type Packet []byte

// New creates a new packet.
// It expects a byteCode for the type of message and
// a data parameter in the form of a byte slice.
// The maximum data size allowed is 65,524 bytes.
// 65535 - 8 (UDP header) - 1 (type) - 2 (length of data)
func New(packetType byte, data []byte) Packet {
	packet := make([]byte, 1+len(data))
	packet[0] = packetType
	copy(packet[1:], data)
	return packet
}

// Type returns the type of the packet.
func (packet Packet) Type() byte {
	return packet[0]
}

// Data returns the data payload.
func (packet Packet) Data() []byte {
	return packet[1:]
}
