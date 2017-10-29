package nano

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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
	data      sync.Map
	dataMutex sync.Mutex
	db        *Database
	name      string
	dirty     chan bool
	close     chan bool
	fileMutex sync.Mutex
	typ       reflect.Type
}

// NewCollection ...
func NewCollection(db *Database, name string) *Collection {
	collection := &Collection{
		db:    db,
		name:  name,
		dirty: make(chan bool, runtime.NumCPU()),
		close: make(chan bool, 1),
	}

	t, exists := collection.db.types[collection.name]

	if !exists {
		panic("Type " + collection.name + " has not been defined")
	}

	collection.typ = t

	if db.IsMaster() {
		collection.loadFromDisk()

		go func() {
			for {
				select {
				case <-collection.dirty:
					for len(collection.dirty) > 0 {
						<-collection.dirty
					}

					collection.flush()
					time.Sleep(db.ioSleepTime)

				case <-collection.close:
					return
				}
			}
		}()
	}

	return collection
}

// Get ...
func (collection *Collection) Get(key string) (interface{}, error) {
	val, ok := collection.data.Load(key)

	if !ok {
		return val, errors.New("Not found")
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

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	if value == nil {
		return
	}

	collection.set(key, value)

	if collection.db.broadcastRequired() {
		var buffer bytes.Buffer

		jsonBytes, err := json.Marshal(value)

		if err != nil {
			panic(err)
		}

		buffer.WriteString(collection.name)
		buffer.WriteByte('\n')
		buffer.WriteString(key)
		buffer.WriteByte('\n')
		buffer.Write(jsonBytes)
		buffer.WriteByte('\n')

		collection.db.broadcast(packet.New(messageSet, buffer.Bytes()))
	}
}

// set ...
func (collection *Collection) set(key string, value interface{}) {
	collection.dataMutex.Lock()
	collection.data.Store(key, value)
	collection.dataMutex.Unlock()

	// The potential data race here does not matter at all.
	if collection.db.IsMaster() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Delete ...
func (collection *Collection) Delete(key string) bool {
	_, exists := collection.data.Load(key)

	collection.dataMutex.Lock()
	collection.data.Delete(key)
	collection.dataMutex.Unlock()

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}

	return exists
}

// Clear deletes all objects from the collection.
func (collection *Collection) Clear() {
	collection.dataMutex.Lock()
	collection.data = sync.Map{}
	collection.dataMutex.Unlock()
	runtime.GC()

	// The potential data race here does not matter at all.
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
		collection.dataMutex.Lock()
		collection.data.Range(func(key, value interface{}) bool {
			channel <- value
			return true
		})
		collection.dataMutex.Unlock()

		close(channel)
	}()

	return channel
}

// flush writes all data to the file system.
func (collection *Collection) flush() {
	collection.fileMutex.Lock()
	defer collection.fileMutex.Unlock()

	file, err := os.OpenFile(path.Join(collection.db.root, collection.name+".dat"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		panic(err)
	}

	file.Seek(0, io.SeekStart)
	bufferedWriter := bufio.NewWriter(file)
	collection.writeRecords(bufferedWriter, true)
	err = bufferedWriter.Flush()

	if err != nil {
		panic(err)
	}

	err = file.Sync()

	if err != nil {
		panic(err)
	}

	err = file.Close()

	if err != nil {
		panic(err)
	}
}

// writeRecords ...
func (collection *Collection) writeRecords(bufferedWriter *bufio.Writer, sorted bool) {
	records := []keyValue{}

	collection.dataMutex.Lock()
	collection.data.Range(func(key, value interface{}) bool {
		records = append(records, keyValue{
			key:   key.(string),
			value: value,
		})
		return true
	})
	collection.dataMutex.Unlock()

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
	filePath := path.Join(collection.db.root, collection.name+".dat")
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

	collection.dataMutex.Lock()
	defer collection.dataMutex.Unlock()

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
			v := reflect.New(collection.typ).Interface()
			json.Unmarshal(value, &v)
			collection.data.Store(key, v)
		}

		count++

		if err != nil {
			break
		}
	}
}

// allValues iterates over all values in a sync.Map and sends them to the given channel.
func allValues(data *sync.Map, channel chan interface{}) {
	data.Range(func(key, value interface{}) bool {
		channel <- value
		return true
	})

	close(channel)
}
