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
func (db *Collection) Get(key string) interface{} {
	val, _ := db.data.Load(key)
	return val
}

// Set ...
func (db *Collection) Set(key string, value interface{}) {
	db.data.Store(key, value)
}

// All ...
func (db *Collection) All() chan interface{} {
	channel := make(chan interface{})

	go allValues(&db.data, channel)

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
