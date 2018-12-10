package nano

const (
	packetCollectionRequest  = iota
	packetCollectionResponse = iota
	packetSet                = iota
	packetDelete             = iota
	packetServerClose        = iota

	// packetPing               = iota
	// packetPong               = iota
	// packetClear              = iota
)
