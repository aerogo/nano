package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// Collection ...
type Collection struct {
	data  sync.Map
	db    *Database
	name  string
	dirty chan bool
}

// NewCollection ...
func NewCollection(db *Database, name string) *Collection {
	collection := &Collection{
		db:    db,
		name:  name,
		dirty: make(chan bool, runtime.NumCPU()),
	}

	go func() {
		for {
			<-collection.dirty

			for len(collection.dirty) > 0 {
				<-collection.dirty
			}

			collection.flush()
			time.Sleep(250 * time.Millisecond)
		}
	}()

	return collection
}

// Get ...
func (collection *Collection) Get(key string) interface{} {
	val, _ := collection.data.Load(key)
	return val
}

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	collection.data.Store(key, value)

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}
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

// flush writes all data to the file system.
func (collection *Collection) flush() {
	start := time.Now()
	file, err := os.OpenFile(collection.name+".dat", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

	if err != nil {
		panic(err)
	}

	file.Seek(0, 0)
	bufferedWriter := bufio.NewWriter(file)

	collection.data.Range(func(key, value interface{}) bool {
		valueBytes, err := json.Marshal(value)

		if err != nil {
			panic(err)
		}

		bufferedWriter.WriteString(key.(string))
		bufferedWriter.WriteByte('\n')

		bufferedWriter.Write(valueBytes)
		bufferedWriter.WriteByte('\n')

		return true
	})

	bufferedWriter.Flush()
	file.Close()
	fmt.Println("flush", time.Since(start))
}

// allValues iterates over all values in a sync.Map and sends them to the given channel.
func allValues(data *sync.Map, channel chan interface{}) {
	data.Range(func(key, value interface{}) bool {
		channel <- value
		return true
	})

	close(channel)
}
