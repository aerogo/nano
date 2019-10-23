package nano

import (
	"os"
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

// newNamespace is the internal function used to create a new namespace.
func newNamespace(node *Node, name string) *Namespace {
	// Create namespace
	namespace := &Namespace{
		node: node,
		name: name,
		root: path.Join(node.config.Directory, name),
	}

	// Create directory
	err := os.MkdirAll(namespace.root, 0777)

	if err != nil {
		panic(err)
	}

	return namespace
}

// RegisterTypes expects a list of pointers and will look up the types
// of the given pointers. These types will be registered so that collections
// can store data using the given type. Note that nil pointers are acceptable.
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

// Exists returns whether or not the key exists.
func (ns *Namespace) Exists(collection string, key string) bool {
	return ns.Collection(collection).Exists(key)
}

// All returns a channel of all objects in the collection.
func (ns *Namespace) All(name string) chan interface{} {
	return ns.Collection(name).All()
}

// Clear deletes all objects from the collection.
func (ns *Namespace) Clear(collection string) {
	ns.Collection(collection).Clear()
}

// ClearAll deletes all objects from all collections,
// effectively resetting the entire database.
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

// Types returns a map of type names mapped to their reflection type.
func (ns *Namespace) Types() map[string]reflect.Type {
	copied := make(map[string]reflect.Type)

	ns.types.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value.(reflect.Type)
		return true
	})

	return copied
}

// HasType returns true if the given type name has been registered.
func (ns *Namespace) HasType(typeName string) bool {
	_, exists := ns.types.Load(typeName)
	return exists
}

// Node returns the cluster node used for this namespace.
func (ns *Namespace) Node() *Node {
	return ns.node
}

// Close will close all collections in the namespace,
// forcing them to sync all data to disk before shutting down.
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
