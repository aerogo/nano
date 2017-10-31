package nano

const (
	packetPing               = iota
	packetPong               = iota
	packetCollectionRequest  = iota
	packetCollectionResponse = iota
	packetSet                = iota
	packetDelete             = iota
	packetClear              = iota
)
