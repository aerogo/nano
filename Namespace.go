package nano

import (
	"os"
	"os/user"
	"path"
	"reflect"
	"sync"
	"time"
)

// Namespace combines multiple collections under a single name.
type Namespace struct {
	collections        sync.Map
	collectionsLoading sync.Map
	name               string
	root               string
	types              sync.Map
	node               *Node
}

// newNamespace ...
func newNamespace(node *Node, name string) *Namespace {
	// Get user info to access the home directory
	user, err := user.Current()

	if err != nil {
		panic(err)
	}

	// Create namespace
	namespace := &Namespace{
		node: node,
		name: name,
		root: path.Join(user.HomeDir, ".aero", "db", name),
	}

	// Create directory
	os.MkdirAll(namespace.root, 0777)

	return namespace
}

// RegisterTypes ...
func (ns *Namespace) RegisterTypes(types ...interface{}) *Namespace {
	// Convert example objects to their respective types
	for _, example := range types {
		typeInfo := reflect.TypeOf(example)

		if typeInfo.Kind() == reflect.Ptr {
			typeInfo = typeInfo.Elem()
		}

		ns.types.Store(typeInfo.Name(), typeInfo)
	}

	return ns
}

// Collection returns the collection with the given name.
func (ns *Namespace) Collection(name string) *Collection {
	obj, loaded := ns.collections.LoadOrStore(name, nil)

	if !loaded {
		collection := newCollection(ns, name)
		ns.collections.Store(name, collection)
		return collection
	}

	// Wait for existing collection load
	for obj == nil {
		time.Sleep(1 * time.Millisecond)
		obj, _ = ns.collections.Load(name)
	}

	return obj.(*Collection)
}

// collectionLoading returns the collection that is currently being loaded.
func (ns *Namespace) collectionLoading(name string) *Collection {
	obj, _ := ns.collectionsLoading.Load(name)
	return obj.(*Collection)
}

// Get returns the value for the given key.
func (ns *Namespace) Get(collection string, key string) (interface{}, error) {
	return ns.Collection(collection).Get(key)
}

// GetMany is the same as Get, except it looks up multiple keys at once.
func (ns *Namespace) GetMany(collection string, keys []string) []interface{} {
	return ns.Collection(collection).GetMany(keys)
}

// Set sets the value for the key.
func (ns *Namespace) Set(collection string, key string, value interface{}) {
	ns.Collection(collection).Set(key, value)
}

// Delete deletes a key from the collection.
func (ns *Namespace) Delete(collection string, key string) bool {
	return ns.Collection(collection).Delete(key)
}

// Exists ...
func (ns *Namespace) Exists(collection string, key string) bool {
	return ns.Collection(collection).Exists(key)
}

// All ...
func (ns *Namespace) All(name string) chan interface{} {
	return ns.Collection(name).All()
}

// Clear ...
func (ns *Namespace) Clear(collection string) {
	ns.Collection(collection).Clear()
}

// ClearAll ...
func (ns *Namespace) ClearAll() {
	ns.collections.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}

		collection := value.(*Collection)
		collection.Clear()

		return true
	})
}

// Types ...
func (ns *Namespace) Types() map[string]reflect.Type {
	copied := make(map[string]reflect.Type)

	ns.types.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value.(reflect.Type)
		return true
	})

	return copied
}

// HasType ...
func (ns *Namespace) HasType(typeName string) bool {
	_, exists := ns.types.Load(typeName)
	return exists
}

// Node ...
func (ns *Namespace) Node() *Node {
	return ns.node
}

// Close ...
func (ns *Namespace) Close() {
	if !ns.node.node.IsServer() {
		return
	}

	ns.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)

		// Stop writing
		collection.close <- true
		<-collection.close

		return true
	})
}

// Prefetch loads all the data for this namespace from disk into memory.
func (ns *Namespace) Prefetch() {
	wg := sync.WaitGroup{}

	ns.types.Range(func(key, value interface{}) bool {
		typeName := key.(string)
		wg.Add(1)

		go func(name string) {
			ns.Collection(name)
			wg.Done()
		}(typeName)

		return true
	})

	wg.Wait()
}
