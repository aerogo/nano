package nano

import (
	"os"
	"os/user"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/aerogo/packet"
)

// Database ...
type Database struct {
	collections sync.Map
	root        string
	ioSleepTime time.Duration
	types       map[string]reflect.Type
	server      Server
	client      Client
}

// New ...
func New(namespace string, types []interface{}) *Database {
	// Convert example objects to their respective types
	collectionTypes := make(map[string]reflect.Type)

	for _, example := range types {
		typeInfo := reflect.TypeOf(example).Elem()
		collectionTypes[typeInfo.Name()] = typeInfo
	}

	// Create directory
	user, err := user.Current()

	if err != nil {
		panic(err)
	}

	root := path.Join(user.HomeDir, ".aero", "db", namespace)
	os.MkdirAll(root, 0777)

	// Create database
	db := &Database{
		root:        root,
		ioSleepTime: 500 * time.Millisecond,
		types:       collectionTypes,
	}

	db.server.db = db
	db.client.db = db
	db.connect()

	return db
}

// Collection ...
func (db *Database) Collection(name string) *Collection {
	obj, found := db.collections.Load(name)

	if !found {
		collection := NewCollection(db, name)
		db.collections.Store(name, collection)
		return collection
	}

	return obj.(*Collection)
}

// Get ...
func (db *Database) Get(collection string, key string) (interface{}, error) {
	return db.Collection(collection).Get(key)
}

// GetMany ...
func (db *Database) GetMany(collection string, keys []string) []interface{} {
	return db.Collection(collection).GetMany(keys)
}

// Set ...
func (db *Database) Set(collection string, key string, value interface{}) {
	db.Collection(collection).Set(key, value)
}

// Delete ...
func (db *Database) Delete(collection string, key string) bool {
	return db.Collection(collection).Delete(key)
}

// Exists ...
func (db *Database) Exists(collection string, key string) bool {
	return db.Collection(collection).Exists(key)
}

// All ...
func (db *Database) All(name string) chan interface{} {
	return db.Collection(name).All()
}

// Clear ...
func (db *Database) Clear(collection string) {
	db.Collection(collection).Clear()
}

// ClearAll ...
func (db *Database) ClearAll() *Database {
	db.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)
		collection.Clear()
		return true
	})

	return db
}

// Types ...
func (db *Database) Types() map[string]reflect.Type {
	return db.types
}

// Close ...
func (db *Database) Close() {
	if db.client.Connection != nil && !db.client.closed {
		db.client.close <- true
	}

	if db.server.listener != nil {
		db.server.close <- true
	}

	// time.Sleep(db.ioSleepTime)

	db.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)

		// We simply try to acquire the lock to assure that any ongoing flush() calls have finished.
		collection.fileMutex.Lock()
		collection.fileMutex.Unlock()
		return true
	})
}

// PrefetchData ...
func (db *Database) PrefetchData() {
	wg := sync.WaitGroup{}

	for typeName := range db.types {
		wg.Add(1)

		go func(name string) {
			db.Collection(name)
			wg.Done()
		}(typeName)
	}

	wg.Wait()
}

// IsMaster ...
func (db *Database) IsMaster() bool {
	return db.server.listener != nil
}

// connect ...
func (db *Database) connect() {
	if db.server.start() != nil {
		db.client.connect()
	}
}

// broadcast ...
func (db *Database) broadcast(msg *packet.Packet) {
	if db.IsMaster() {
		db.server.broadcasts <- msg
	} else {
		db.client.Outgoing <- msg
	}
}

// broadcastRequired ...
func (db *Database) broadcastRequired() bool {
	return db.client.Connection != nil || len(db.server.connections) > 0
}
