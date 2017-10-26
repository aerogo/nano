package database

import (
	"sync"
)

// Database ...
type Database struct {
	collections sync.Map
	root        string
}

// New ...
func New(root string) *Database {
	return &Database{
		root: root,
	}
}

// Close ...
func (db *Database) Close() {
	db.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)
		collection.flush()

		return true
	})
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
func (db *Database) Get(collection string, key string) interface{} {
	return db.Collection(collection).Get(key)
}

// Set ...
func (db *Database) Set(collection string, key string, value interface{}) {
	db.Collection(collection).Set(key, value)
}

// Delete ...
func (db *Database) Delete(collection string, key string) {
	db.Collection(collection).Delete(key)
}

// Exists ...
func (db *Database) Exists(collection string, key string) bool {
	return db.Collection(collection).Exists(key)
}

// All ...
func (db *Database) All(name string) chan interface{} {
	return db.Collection(name).All()
}
