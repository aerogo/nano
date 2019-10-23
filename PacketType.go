package nano

const (
	packetCollectionRequest  = iota
	packetCollectionResponse = iota
	packetSet                = iota
	packetDelete             = iota
	packetServerClose        = iota
)
