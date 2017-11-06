package nano

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/aerogo/packet"
)

// ChannelBufferSize is the size of the channels used to iterate over a whole collection.
const ChannelBufferSize = 128

// Collection ...
type Collection struct {
	data             sync.Map
	lastModification sync.Map
	ns               *Namespace
	node             *Node
	name             string
	dirty            chan bool
	close            chan bool
	loaded           chan bool
	fileMutex        sync.Mutex
	typ              reflect.Type
}

// newCollection ...
func newCollection(ns *Namespace, name string) *Collection {
	collection := &Collection{
		ns:     ns,
		node:   ns.node,
		name:   name,
		dirty:  make(chan bool, runtime.NumCPU()),
		close:  make(chan bool),
		loaded: make(chan bool),
	}

	t, exists := collection.ns.types[collection.name]

	if !exists {
		panic("Type " + collection.name + " has not been defined")
	}

	collection.typ = t
	collection.load()

	return collection
}

// load loads all collection data
func (collection *Collection) load() {
	if collection.node.IsServer() {
		// Server
		collection.loadFromDisk()

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
						collection.flush()
					}

					close(collection.close)
					return
				}
			}
		}()
	} else {
		// Client
		collection.ns.collectionsLoading.Store(collection.name, collection)
		data := fmt.Sprintf("%s\n%s\n", collection.ns.name, collection.name)
		collection.node.Client().Stream.Outgoing <- packet.New(packetCollectionRequest, []byte(data))
		<-collection.loaded
	}
}

// Get ...
func (collection *Collection) Get(key string) (interface{}, error) {
	val, ok := collection.data.Load(key)

	if !ok {
		return val, errors.New("Key not found: " + key)
	}

	return val, nil
}

// GetMany ...
func (collection *Collection) GetMany(keys []string) []interface{} {
	values := make([]interface{}, len(keys), len(keys))

	for i := 0; i < len(keys); i++ {
		values[i], _ = collection.Get(keys[i])
	}

	return values
}

// set ...
func (collection *Collection) set(key string, value interface{}) {
	collection.data.Store(key, value)

	if collection.node.IsServer() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	if value == nil {
		return
	}

	if collection.node.broadcastRequired() {
		// It's important to store the timestamp BEFORE the actual collection.set
		collection.lastModification.Store(key, time.Now().UnixNano())

		jsonBytes, err := json.Marshal(value)

		if err != nil {
			panic(err)
		}

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

// delete ...
func (collection *Collection) delete(key string) {
	collection.data.Delete(key)

	if collection.node.IsServer() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Delete ...
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
}

// Exists ...
func (collection *Collection) Exists(key string) bool {
	_, exists := collection.data.Load(key)
	return exists
}

// All ...
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

// flush writes all data to the file system.
func (collection *Collection) flush() error {
	collection.fileMutex.Lock()
	defer collection.fileMutex.Unlock()

	file, err := os.OpenFile(path.Join(collection.ns.root, collection.name+".dat"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	file.Seek(0, io.SeekStart)
	bufferedWriter := bufio.NewWriter(file)
	collection.writeRecords(bufferedWriter, true)
	err = bufferedWriter.Flush()

	if err != nil {
		return err
	}

	err = file.Sync()

	if err != nil {
		return err
	}

	return file.Close()
}

// writeRecords ...
func (collection *Collection) writeRecords(bufferedWriter *bufio.Writer, sorted bool) {
	records := []keyValue{}

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

	for _, record := range records {
		valueBytes, err := json.Marshal(record.value)

		if err != nil {
			panic(err)
		}

		bufferedWriter.WriteString(record.key)
		bufferedWriter.WriteByte('\n')

		bufferedWriter.Write(valueBytes)
		bufferedWriter.WriteByte('\n')
	}
}

// loadFromDisk ...
func (collection *Collection) loadFromDisk() {
	filePath := path.Join(collection.ns.root, collection.name+".dat")
	stream, err := os.OpenFile(filePath, os.O_RDONLY|os.O_SYNC, 0644)

	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		panic(err)
	}

	collection.readRecords(stream)
}

// readRecords ...
func (collection *Collection) readRecords(stream io.Reader) {
	var key string
	var value []byte

	reader := bufio.NewReader(stream)
	count := 0

	for {
		line, err := reader.ReadBytes('\n')

		// Remove delimiter
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if count%2 == 0 {
			key = string(line)
		} else {
			value = line
			v := reflect.New(collection.typ)
			obj := v.Interface()
			json.Unmarshal(value, &obj)
			collection.data.Store(key, obj)
		}

		count++

		if err != nil {
			break
		}
	}
}
