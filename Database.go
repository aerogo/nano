package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Database ...
type Database struct {
	collections sync.Map
	root        string
	ioSleepTime time.Duration
	types       map[string]reflect.Type
}

// New ...
func New(root string, types []interface{}) *Database {
	// Convert example objects to their respective types
	collectionTypes := make(map[string]reflect.Type)

	for _, example := range types {
		typeInfo := reflect.TypeOf(example).Elem()
		collectionTypes[typeInfo.Name()] = typeInfo
	}

	// Create database
	db := &Database{
		root:        root,
		ioSleepTime: 500 * time.Millisecond,
		types:       collectionTypes,
	}

	// Load existing date from disk
	start := time.Now()
	db.loadFiles()
	fmt.Println(time.Since(start))

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

// Close ...
func (db *Database) Close() {
	db.collections.Range(func(key, value interface{}) bool {
		collection := value.(*Collection)
		collection.flush()
		return true
	})
}

// loadFiles ...
func (db *Database) loadFiles() {
	files, err := ioutil.ReadDir(db.root)

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") || !strings.HasSuffix(file.Name(), ".dat") {
			continue
		}

		collectionName := strings.TrimSuffix(file.Name(), ".dat")

		t, exists := db.types[collectionName]

		if !exists {
			panic("Type " + collectionName + " has not been defined")
		}

		stream, err := os.OpenFile(path.Join(db.root, file.Name()), os.O_RDONLY, 0666)

		if err != nil {
			panic(err)
		}

		collection := db.Collection(collectionName)

		var key string
		var value []byte

		scanner := bufio.NewScanner(stream)
		count := 0

		for scanner.Scan() {
			if count%2 == 0 {
				key = scanner.Text()
			} else {
				value = scanner.Bytes()
				v := reflect.New(t).Interface()
				json.Unmarshal(value, &v)
				collection.data.Store(key, v)
			}

			count++
		}

		err = scanner.Err()

		if err != nil {
			panic(err)
		}
	}
}
