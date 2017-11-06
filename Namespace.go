package nano

import (
	"os"
	"os/user"
	"path"
	"reflect"
	"sync"
	"time"
)

// Namespace ...
type Namespace struct {
	collections        sync.Map
	collectionsLoading sync.Map
	name               string
	root               string
	types              map[string]reflect.Type
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
		node:  node,
		name:  name,
		root:  path.Join(user.HomeDir, ".aero", "db", name),
		types: make(map[string]reflect.Type),
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

		ns.types[typeInfo.Name()] = typeInfo
	}

	return ns
}

// Collection ...
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

// Get ...
func (ns *Namespace) Get(collection string, key string) (interface{}, error) {
	return ns.Collection(collection).Get(key)
}

// GetMany ...
func (ns *Namespace) GetMany(collection string, keys []string) []interface{} {
	return ns.Collection(collection).GetMany(keys)
}

// Set ...
func (ns *Namespace) Set(collection string, key string, value interface{}) {
	ns.Collection(collection).Set(key, value)
}

// Delete ...
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
	return ns.types
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

// Prefetch ...
func (ns *Namespace) Prefetch() {
	wg := sync.WaitGroup{}

	for typeName := range ns.types {
		wg.Add(1)

		go func(name string) {
			ns.Collection(name)
			wg.Done()
		}(typeName)
	}

	wg.Wait()
}
