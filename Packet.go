package nano

// Packet ...
type Packet struct {
	Type   byte
	Length int64
	Data   []byte
}

// NewPacket ...
func NewPacket(byteCode byte, data []byte) *Packet {
	return &Packet{
		Type:   byteCode,
		Length: int64(len(data)),
		Data:   data,
	}
}

// Bytes ...
func (packet *Packet) Bytes() []byte {
	result := []byte{packet.Type}
	result = append(result, toBytes(packet.Length)...)
	result = append(result, packet.Data...)
	return result
}
