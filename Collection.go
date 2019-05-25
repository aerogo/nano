package nano

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aerogo/packet"
	jsoniter "github.com/json-iterator/go"
)

// ChannelBufferSize is the size of the channels used to iterate over a whole collection.
const ChannelBufferSize = 128

// Collection is a hash map of data of the same type that is synchronized across network and disk.
type Collection struct {
	data             sync.Map
	lastModification sync.Map
	ns               *Namespace
	node             *Node
	name             string
	dirty            chan bool
	close            chan bool
	loaded           chan bool
	count            int64
	fileMutex        sync.Mutex
	typ              reflect.Type
}

// newCollection creates a new collection in the namespace with the given name.
func newCollection(ns *Namespace, name string) *Collection {
	collection := &Collection{
		ns:     ns,
		node:   ns.node,
		name:   name,
		dirty:  make(chan bool, runtime.NumCPU()),
		close:  make(chan bool),
		loaded: make(chan bool),
	}

	t, exists := collection.ns.types.Load(collection.name)

	if !exists {
		panic("Type " + collection.name + " has not been defined")
	}

	collection.typ = t.(reflect.Type)
	collection.load()

	return collection
}

// load loads all collection data
func (collection *Collection) load() {
	if collection.node.IsServer() {
		// Server loads the collection from disk
		err := collection.loadFromDisk()

		if err != nil {
			panic(err)
		}

		// Indicate that collection is loaded
		close(collection.loaded)

		go func() {
			for {
				select {
				case <-collection.dirty:
					for len(collection.dirty) > 0 {
						<-collection.dirty
					}

					err := collection.flush()

					if err != nil {
						fmt.Println("Error writing collection", collection.name, "to disk", err)
					}

					time.Sleep(collection.node.ioSleepTime)

				case <-collection.close:
					if len(collection.dirty) > 0 {
						err := collection.flush()

						if err != nil {
							fmt.Println("Error writing collection", collection.name, "to disk", err)
						}
					}

					close(collection.close)
					return
				}
			}
		}()
	} else {
		// Client asks the server to send the most recent collection data
		collection.ns.collectionsLoading.Store(collection.name, collection)
		packetData := bytes.Buffer{}
		fmt.Fprintf(&packetData, "%s\n%s\n", collection.ns.name, collection.name)
		collection.node.Client().Stream.Outgoing <- packet.New(packetCollectionRequest, packetData.Bytes())
		<-collection.loaded
	}
}

// Get returns the value for the given key.
func (collection *Collection) Get(key string) (interface{}, error) {
	val, ok := collection.data.Load(key)

	if !ok {
		return val, errors.New("Key not found: " + key)
	}

	return val, nil
}

// GetMany is the same as Get, except it looks up multiple keys at once.
func (collection *Collection) GetMany(keys []string) []interface{} {
	values := make([]interface{}, len(keys))

	for i := 0; i < len(keys); i++ {
		values[i], _ = collection.Get(keys[i])
	}

	return values
}

// set is the internally used function to store a value for a key.
func (collection *Collection) set(key string, value interface{}) {
	collection.data.Store(key, value)

	if collection.node.IsServer() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Set sets the value for the key.
func (collection *Collection) Set(key string, value interface{}) {
	if value == nil {
		return
	}

	if collection.node.broadcastRequired() {
		// It's important to store the timestamp BEFORE the actual collection.set
		collection.lastModification.Store(key, time.Now().UnixNano())

		// Serialize the value into JSON format
		jsonBytes, err := jsoniter.Marshal(value)

		if err != nil {
			panic(err)
		}

		// Create a network packet for the "set" command
		buffer := bytes.Buffer{}
		buffer.Write(packet.Int64ToBytes(time.Now().UnixNano()))
		buffer.WriteString(collection.ns.name)
		buffer.WriteByte('\n')
		buffer.WriteString(collection.name)
		buffer.WriteByte('\n')
		buffer.WriteString(key)
		buffer.WriteByte('\n')
		buffer.Write(jsonBytes)
		buffer.WriteByte('\n')

		msg := packet.New(packetSet, buffer.Bytes())
		collection.node.Broadcast(msg)
	}

	collection.set(key, value)
}

// delete is the internally used command to delete a key.
func (collection *Collection) delete(key string) {
	collection.data.Delete(key)

	if collection.node.IsServer() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Delete deletes a key from the collection.
func (collection *Collection) Delete(key string) bool {
	if collection.node.broadcastRequired() {
		// It's important to store the timestamp BEFORE the actual collection.delete
		collection.lastModification.Store(key, time.Now().UnixNano())

		buffer := bytes.Buffer{}
		buffer.Write(packet.Int64ToBytes(time.Now().UnixNano()))
		buffer.WriteString(collection.ns.name)
		buffer.WriteByte('\n')
		buffer.WriteString(collection.name)
		buffer.WriteByte('\n')
		buffer.WriteString(key)
		buffer.WriteByte('\n')

		msg := packet.New(packetDelete, buffer.Bytes())
		collection.node.Broadcast(msg)
	}

	_, exists := collection.data.Load(key)
	collection.delete(key)

	return exists
}

// Clear deletes all objects from the collection.
func (collection *Collection) Clear() {
	collection.data.Range(func(key, value interface{}) bool {
		collection.data.Delete(key)
		return true
	})

	runtime.GC()

	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}

	atomic.StoreInt64(&collection.count, 0)
}

// Exists returns whether or not the key exists.
func (collection *Collection) Exists(key string) bool {
	_, exists := collection.data.Load(key)
	return exists
}

// All returns a channel of all objects in the collection.
func (collection *Collection) All() chan interface{} {
	channel := make(chan interface{}, ChannelBufferSize)

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
func (collection *Collection) Count() int64 {
	return atomic.LoadInt64(&collection.count)
}

// flush writes all data to the file system.
func (collection *Collection) flush() error {
	collection.fileMutex.Lock()
	defer collection.fileMutex.Unlock()

	newFilePath := path.Join(collection.ns.root, collection.name+".new")
	oldFilePath := path.Join(collection.ns.root, collection.name+".dat")
	tmpFilePath := path.Join(collection.ns.root, collection.name+".tmp")

	file, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	bufferedWriter := bufio.NewWriter(file)
	err = collection.writeRecords(bufferedWriter, true)

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

	// Swap .dat and .new files
	err = os.Rename(oldFilePath, tmpFilePath)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.Rename(newFilePath, oldFilePath)

	if err != nil {
		return err
	}

	err = os.Remove(tmpFilePath)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// writeRecords writes the entire collection to the IO writer.
func (collection *Collection) writeRecords(writer io.Writer, sorted bool) error {
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
	encoder := jsoniter.NewEncoder(writer)

	for _, record := range records {
		// Key in the first line
		_, err := stringWriter.WriteString(record.key)

		if err != nil {
			return err
		}

		_, err = stringWriter.WriteString("\n")

		if err != nil {
			return err
		}

		// Value in the second line
		err = encoder.Encode(record.value)

		if err != nil {
			return err
		}
	}

	return nil
}

// loadFromDisk loads the entire collection from disk.
func (collection *Collection) loadFromDisk() error {
	filePath := path.Join(collection.ns.root, collection.name+".dat")
	stream, err := os.OpenFile(filePath, os.O_RDONLY|os.O_SYNC, 0644)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return collection.readRecords(stream)
}

// readRecords reads the entire collection from an IO reader.
func (collection *Collection) readRecords(stream io.Reader) error {
	var key string
	var value []byte

	reader := bufio.NewReader(stream)
	lineCount := 0

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

		// Remove delimiter
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if lineCount%2 == 0 {
			key = string(line)
		} else {
			value = line
			v := reflect.New(collection.typ)
			obj := v.Interface()
			err = jsoniter.Unmarshal(value, &obj)

			if err != nil {
				return err
			}

			collection.data.Store(key, obj)
		}

		lineCount++
	}
}
