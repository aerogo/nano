package packet

const (
	// BufferSize is set to the maximum length of 65,535 bytes.
	BufferSize = 65535

	// MaxDataLength is set to the maximum data length of 65,524 bytes.
	// It equals BufferSize - 8 (UDP header) - 1 (type) - 2 (length of data).
	MaxDataLength = BufferSize - 8 - 1 - 2
)

// Packet is a byte slice abstraction to send types of messages
// within a packet.
type Packet []byte

// Type is a bytecode that describes the type of the message.
type Type byte

// New creates a new packet.
// It expects a byteCode for the type of message and
// a data parameter in the form of a byte slice.
// The maximum data size allowed is specified in MaxDataLength.
func New(packetType Type, data []byte) Packet {
	packet := make([]byte, 1+len(data))
	packet[0] = byte(packetType)
	copy(packet[1:], data)
	return packet
}

// Type returns the type of the packet.
func (packet Packet) Type() Type {
	return Type(packet[0])
}

// Data returns the data payload.
func (packet Packet) Data() []byte {
	return packet[1:]
}
