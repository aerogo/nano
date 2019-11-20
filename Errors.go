package nano

import "errors"

var (
	// ErrKeyNotFound is returned when the key was not found in the collection.
	ErrKeyNotFound = errors.New("Key not found")
)
