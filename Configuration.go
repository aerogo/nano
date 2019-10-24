package nano

// Configuration represents the nano configuration
// which is only read once at node creation time.
type Configuration struct {
	// Port is the port used by the server and client nodes.
	Port int

	// Directory includes the path to the namespaces stored on the disk.
	Directory string

	// Hosts represents a whitelist of addresses allowed to access the database.
	Hosts []string
}
