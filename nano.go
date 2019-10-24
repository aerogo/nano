package nano

// New creates a new nano node.
func New(config Configuration) Node {
	node := &node{
		config: config,
	}

	err := node.connect()

	if err != nil {
		panic(err)
	}

	return node
}
