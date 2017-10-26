package database

import (
	"encoding/json"
	"os"
	"sync"
)

// Collection ...
type Collection struct {
	data          sync.Map
	db            *Database
	file          *os.File
	fileMutex     sync.Mutex
	currentOffset int64
	offsets       map[string]int64
}

// NewCollection ...
func NewCollection(db *Database, name string) *Collection {
	file, err := os.OpenFile(name+".dat", os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		panic(err)
	}

	return &Collection{
		db:   db,
		file: file,
	}
}

// Get ...
func (collection *Collection) Get(key string) interface{} {
	val, _ := collection.data.Load(key)
	return val
}

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	collection.data.Store(key, value)

	go func() {
		collection.fileMutex.Lock()
		valueBytes, err := json.Marshal(value)

		if err != nil {
			panic(err)
		}

		valueBytes = append(valueBytes, '\n')
		written, err := collection.file.WriteAt(append([]byte(key+"\n"), valueBytes...), collection.currentOffset)

		if err != nil {
			panic(err)
		}

		collection.currentOffset += int64(written)
		collection.fileMutex.Unlock()
	}()
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
