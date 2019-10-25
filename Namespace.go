package nano

import (
	"os"
	"path"
	"reflect"
	"sync"
)

// Namespace combines multiple collections under a single name.
type Namespace interface {
	All(name string) chan interface{}
	Clear(collection string)
	ClearAll()
	Close()
	Delete(collection string, key string) bool
	Exists(collection string, key string) bool
	Get(collection string, key string) (interface{}, error)
	GetMany(collection string, keys []string) []interface{}
	HasType(typeName string) bool
	Node() Node
	Prefetch()
	Set(collection string, key string, value interface{})
	Types() map[string]reflect.Type
}

type namespace struct {
	collections sync.Map
	types       sync.Map
	name        string
	root        string
	node        *node
}

// newNamespace is the internal function used to create a new namespace.
func newNamespace(node *node, name string) (*namespace, error) {
	namespace := &namespace{
		node: node,
		name: name,
		root: path.Join(node.config.Directory, name),
	}

	err := os.MkdirAll(namespace.root, 0777)

	if err != nil {
		return nil, err
	}

	return namespace, nil
}

// All returns a channel of all objects in the collection.
func (ns *namespace) All(name string) chan interface{} {
	return ns.Collection(name).All()
}

// Clear deletes all objects from the collection.
func (ns *namespace) Clear(collection string) {
	ns.Collection(collection).Clear()
}

// ClearAll deletes all objects from all collections,
// effectively resetting the entire database.
func (ns *namespace) ClearAll() {
	ns.collections.Range(func(key, value interface{}) bool {
		collection := value.(Collection)
		collection.Clear()
		return true
	})
}

// Close will close all collections in the namespace,
// forcing them to sync all data to disk before shutting down.
func (ns *namespace) Close() {
	ns.collections.Range(func(key, value interface{}) bool {
		collection := value.(Collection)
		collection.Close()
		return true
	})
}

// RegisterTypes expects a list of pointers and will look up the types
// of the given pointers. These types will be registered so that collections
// can store data using the given type. Note that nil pointers are acceptable.
func (ns *namespace) RegisterTypes(types ...interface{}) *namespace {
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
func (ns *namespace) Collection(name string) Collection {
	obj, loaded := ns.collections.LoadOrStore(name, nil)

	if !loaded {
		collection := newCollection(ns, name)
		ns.collections.Store(name, collection)
		return collection
	}

	return obj.(Collection)
}

// Delete deletes a key from the collection.
func (ns *namespace) Delete(collection string, key string) bool {
	return ns.Collection(collection).Delete(key)
}

// Exists returns whether or not the key exists.
func (ns *namespace) Exists(collection string, key string) bool {
	return ns.Collection(collection).Exists(key)
}

// Get returns the value for the given key.
func (ns *namespace) Get(collection string, key string) (interface{}, error) {
	return ns.Collection(collection).Get(key)
}

// GetMany is the same as Get, except it looks up multiple keys at once.
func (ns *namespace) GetMany(collection string, keys []string) []interface{} {
	return ns.Collection(collection).GetMany(keys)
}

// HasType returns true if the given type name has been registered.
func (ns *namespace) HasType(typeName string) bool {
	_, exists := ns.types.Load(typeName)
	return exists
}

// Node returns the cluster node used for this namespace.
func (ns *namespace) Node() Node {
	return ns.node
}

// Prefetch loads all the data for this namespace from disk into memory.
func (ns *namespace) Prefetch() {
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

// Set sets the value for the key.
func (ns *namespace) Set(collection string, key string, value interface{}) {
	ns.Collection(collection).Set(key, value)
}

// Types returns a map of type names mapped to their reflection type.
func (ns *namespace) Types() map[string]reflect.Type {
	copied := make(map[string]reflect.Type)

	ns.types.Range(func(key, value interface{}) bool {
		copied[key.(string)] = value.(reflect.Type)
		return true
	})

	return copied
}
