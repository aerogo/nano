package nano

import (
	"os"
	"os/user"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/aerogo/cluster"
	"github.com/aerogo/cluster/client"
	"github.com/aerogo/cluster/server"
)

// Database ...
type Database struct {
	collections sync.Map
	root        string
	ioSleepTime time.Duration
	types       map[string]reflect.Type
	node        cluster.Node
	server      *server.Node
	client      *client.Node
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

// Node ...
func (db *Database) Node() cluster.Node {
	return db.node
}

// Close ...
func (db *Database) Close() {
	db.node.Close()

	db.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)

		if db.node.IsServer() {
			collection.close <- true
		}

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

// connect ...
func (db *Database) connect() {
	db.node = cluster.New(3000)

	if db.node.IsServer() {
		db.server = db.node.(*server.Node)
		db.server.OnConnect(onConnect(db))
	} else {
		db.client = db.node.(*client.Node)
		go clientReadPackets(db.client, db)
	}
}

// broadcastRequired ...
func (db *Database) broadcastRequired() bool {
	if !db.node.IsServer() {
		return true
	}

	return db.server.ClientCount() > 0
}
