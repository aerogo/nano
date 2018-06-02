package nano

const (
	packetCollectionRequest  = iota
	packetCollectionResponse = iota
	packetSet                = iota
	packetDelete             = iota
	packetClose              = iota

	// packetPing               = iota
	// packetPong               = iota
	// packetClear              = iota
)
