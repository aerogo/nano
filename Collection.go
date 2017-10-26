package database

import (
	"sync"
)

// Collection ...
type Collection struct {
	data sync.Map
}

// NewCollection ...
func NewCollection() *Collection {
	return &Collection{}
}

// Get ...
func (collection *Collection) Get(key string) interface{} {
	val, _ := collection.data.Load(key)
	return val
}

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	collection.data.Store(key, value)
}

// Delete ...
func (collection *Collection) Delete(key string) {
	collection.data.Delete(key)
}

// Exists ...
func (collection *Collection) Exists(key string) bool {
	_, exists := collection.data.Load(key)
	return exists
}

// All ...
func (collection *Collection) All() chan interface{} {
	channel := make(chan interface{})

	go allValues(&collection.data, channel)

	return channel
}

// allValues iterates over all values in a sync.Map and sends them to the given channel.
func allValues(data *sync.Map, channel chan interface{}) {
	data.Range(func(key, value interface{}) bool {
		channel <- value
		return true
	})

	close(channel)
}
