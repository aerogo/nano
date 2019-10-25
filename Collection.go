package nano

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// Collection is a hash map of data of the same type that is synchronized across network and disk.
type Collection interface {
	All() chan interface{}
	Clear()
	Close()
	Count() int64
	Delete(key string) bool
	Exists(key string) bool
	Get(key string) (interface{}, error)
	GetMany(keys []string) []interface{}
	Set(key string, value interface{})
}

type collection struct {
	data             sync.Map
	lastModification sync.Map
	name             string
	ns               *namespace
	node             *node
	dirty            chan struct{}
	close            chan struct{}
	count            int64
	flushMutex       sync.Mutex
	typ              reflect.Type
}

// newCollection creates a new collection in the namespace with the given name.
func newCollection(ns *namespace, name string) *collection {
	collection := &collection{
		ns:    ns,
		node:  ns.node,
		name:  name,
		dirty: make(chan struct{}),
		close: make(chan struct{}),
	}

	t, exists := collection.ns.types.Load(collection.name)

	if !exists {
		panic("Type " + collection.name + " has not been defined")
	}

	collection.typ = t.(reflect.Type)
	err := collection.loadFromDisk()

	if err != nil {
		panic(err)
	}

	go collection.writeToDisk()
	return collection
}

// All returns a channel of all objects in the collection.
func (collection *collection) All() chan interface{} {
	channel := make(chan interface{})

	go func() {
		collection.data.Range(func(key, value interface{}) bool {
			channel <- value
			return true
		})

		close(channel)
	}()

	return channel
}

// Count gives you a rough estimate of how many elements are in the collection.
// It DOES NOT GUARANTEE that the returned number is the actual number of elements.
// A good use for this function is to preallocate slices with the given capacity.
// In the future, this function could possibly return the exact number of elements.
func (collection *collection) Count() int64 {
	return atomic.LoadInt64(&collection.count)
}

// Close terminates the syncing goroutine and waits for all flush calls to finish.
func (collection *collection) Close() {
	close(collection.close)

	// Wait for flush calls to finish.
	// nolint:staticcheck
	{
		collection.flushMutex.Lock()
		collection.flushMutex.Unlock()
	}
}

// Get returns the value for the given key.
func (collection *collection) Get(key string) (interface{}, error) {
	value, ok := collection.data.Load(key)

	if !ok {
		return value, errors.New("Key not found: " + key)
	}

	return value, nil
}

// GetMany is the same as Get, except it looks up multiple keys at once.
func (collection *collection) GetMany(keys []string) []interface{} {
	values := make([]interface{}, len(keys))

	for i := 0; i < len(keys); i++ {
		values[i], _ = collection.Get(keys[i])
	}

	return values
}

// set is the internally used function to store a value for a key.
func (collection *collection) set(key string, value interface{}) {
	// It's important to store the timestamp BEFORE storing the data
	collection.lastModification.Store(key, time.Now().UnixNano())
	collection.data.Store(key, value)

	// if collection.node.IsServer() {
	// 	collection.dirty <- struct{}{}
	// }
}

// Set sets the value for the key.
func (collection *collection) Set(key string, value interface{}) {
	if value == nil {
		return
	}

	// if collection.node.broadcastRequired() {
	// 	// Serialize the value into JSON format
	// 	jsonBytes, err := jsoniter.Marshal(value)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// Create a network packet for the "set" command
	// 	buffer := bytes.Buffer{}
	// 	buffer.Write(packet.Int64ToBytes(time.Now().UnixNano()))
	// 	buffer.WriteString(collection.ns.name)
	// 	buffer.WriteByte('\n')
	// 	buffer.WriteString(collection.name)
	// 	buffer.WriteByte('\n')
	// 	buffer.WriteString(key)
	// 	buffer.WriteByte('\n')
	// 	buffer.Write(jsonBytes)
	// 	buffer.WriteByte('\n')

	// 	msg := packet.New(packetSet, buffer.Bytes())
	// 	collection.node.Broadcast(msg)
	// }

	collection.set(key, value)
}

// delete is the internally used command to delete a key.
func (collection *collection) delete(key string) {
	collection.data.Delete(key)

	// if collection.node.IsServer() && len(collection.dirty) == 0 {
	// 	collection.dirty <- true
	// }
}

// Delete deletes a key from the collection.
func (collection *collection) Delete(key string) bool {
	// if collection.node.broadcastRequired() {
	// 	// It's important to store the timestamp BEFORE the actual collection.delete
	// 	collection.lastModification.Store(key, time.Now().UnixNano())

	// 	buffer := bytes.Buffer{}
	// 	buffer.Write(packet.Int64ToBytes(time.Now().UnixNano()))
	// 	buffer.WriteString(collection.ns.name)
	// 	buffer.WriteByte('\n')
	// 	buffer.WriteString(collection.name)
	// 	buffer.WriteByte('\n')
	// 	buffer.WriteString(key)
	// 	buffer.WriteByte('\n')

	// 	msg := packet.New(packetDelete, buffer.Bytes())
	// 	collection.node.Broadcast(msg)
	// }

	_, exists := collection.data.Load(key)
	collection.delete(key)
	return exists
}

// Clear deletes all objects from the collection.
func (collection *collection) Clear() {
	collection.data.Range(func(key, value interface{}) bool {
		collection.data.Delete(key)
		return true
	})

	// if len(collection.dirty) == 0 {
	// 	collection.dirty <- struct{}{}
	// }

	atomic.StoreInt64(&collection.count, 0)
}

// Exists returns whether or not the key exists.
func (collection *collection) Exists(key string) bool {
	_, exists := collection.data.Load(key)
	return exists
}

// flush writes all data to the file system.
func (collection *collection) flush() error {
	collection.flushMutex.Lock()
	defer collection.flushMutex.Unlock()

	newFilePath := path.Join(os.TempDir(), collection.name+".new")
	oldFilePath := path.Join(collection.ns.root, collection.name+".dat")

	file, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	bufferedWriter := bufio.NewWriter(file)
	err = collection.writeTo(bufferedWriter, true)

	if err != nil {
		return err
	}

	err = bufferedWriter.Flush()

	if err != nil {
		return err
	}

	err = file.Sync()

	if err != nil {
		return err
	}

	err = file.Close()

	if err != nil {
		return err
	}

	err = os.Rename(newFilePath, oldFilePath)

	if err != nil {
		return err
	}

	return nil
}

// writeTo writes the entire collection to the IO writer.
func (collection *collection) writeTo(writer io.Writer, sorted bool) error {
	records := []keyValue{}
	stringWriter, ok := writer.(io.StringWriter)

	if !ok {
		return errors.New("The given io.Writer is not an io.StringWriter")
	}

	collection.data.Range(func(key, value interface{}) bool {
		records = append(records, keyValue{
			key:   key.(string),
			value: value,
		})
		return true
	})

	if sorted {
		sort.Slice(records, func(i, j int) bool {
			return records[i].key < records[j].key
		})
	}

	atomic.StoreInt64(&collection.count, int64(len(records)))
	encoder := json.NewEncoder(writer)

	for _, record := range records {
		// 1st line: Key
		_, err := stringWriter.WriteString(record.key)

		if err != nil {
			return err
		}

		_, err = stringWriter.WriteString("\n")

		if err != nil {
			return err
		}

		// 2nd line: Value
		err = encoder.Encode(record.value)

		if err != nil {
			return err
		}
	}

	return nil
}

// writeToDisk writes the entire collection to disk.
func (collection *collection) writeToDisk() {
	defer fmt.Println("collection.writeToDisk shutdown")

	for {
		select {
		case <-collection.dirty:
			err := collection.flush()

			if err != nil {
				fmt.Println("Error writing collection", collection.name, "to disk", err)
			}

			time.Sleep(diskWriteDelay)

		case <-collection.close:
			if len(collection.dirty) > 0 {
				err := collection.flush()

				if err != nil {
					fmt.Println("Error writing collection", collection.name, "to disk", err)
				}
			}

			return
		}
	}
}

// loadFromDisk loads the entire collection from disk.
func (collection *collection) loadFromDisk() error {
	filePath := path.Join(collection.ns.root, collection.name+".dat")
	stream, err := os.OpenFile(filePath, os.O_RDONLY|os.O_SYNC, 0644)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return collection.readFrom(stream)
}

// readFrom reads the entire collection from an IO reader.
func (collection *collection) readFrom(stream io.Reader) error {
	var (
		key       string
		reader    = bufio.NewReader(stream)
		lineCount = 0
	)

	defer func() {
		atomic.StoreInt64(&collection.count, int64(lineCount/2))
	}()

	for {
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if lineCount%2 == 0 {
			key = string(line)
		} else {
			v := reflect.New(collection.typ)
			obj := v.Interface()
			err = jsoniter.Unmarshal(line, &obj)

			if err != nil {
				return err
			}

			collection.data.Store(key, obj)
		}

		lineCount++
	}
}
