package nano

import (
	"bufio"
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
)

// ChannelBufferSize is the size of the channels used to iterate over a whole collection.
const ChannelBufferSize = 128

// Collection ...
type Collection struct {
	data      sync.Map
	db        *Database
	name      string
	dirty     chan bool
	fileMutex sync.Mutex
}

// NewCollection ...
func NewCollection(db *Database, name string) *Collection {
	collection := &Collection{
		db:    db,
		name:  name,
		dirty: make(chan bool, runtime.NumCPU()),
	}

	if db.IsMaster() {
		collection.loadFromDisk()

		go func() {
			for {
				<-collection.dirty

				for len(collection.dirty) > 0 {
					<-collection.dirty
				}

				collection.flush()
				time.Sleep(db.ioSleepTime)
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

	collection.data.Store(key, value)

	// The potential data race here does not matter at all.
	if collection.db.IsMaster() && len(collection.dirty) == 0 {
		collection.dirty <- true
	}

	// collection.db.broadcast(packet.New(messageSet, key+"\n"+))
}

// Delete ...
func (collection *Collection) Delete(key string) bool {
	_, exists := collection.data.Load(key)
	collection.data.Delete(key)

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}

	return exists
}

// Clear deletes all objects from the collection.
func (collection *Collection) Clear() {
	collection.data = sync.Map{}
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

	go allValues(&collection.data, channel)

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

	records := []keyValue{}

	collection.data.Range(func(key, value interface{}) bool {
		records = append(records, keyValue{
			key:   key.(string),
			value: value,
		})
		return true
	})

	sort.Slice(records, func(i, j int) bool {
		return records[i].key < records[j].key
	})

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

// loadFromDisk ...
func (collection *Collection) loadFromDisk() {
	t, exists := collection.db.types[collection.name]

	if !exists {
		panic("Type " + collection.name + " has not been defined")
	}

	filePath := path.Join(collection.db.root, collection.name+".dat")
	stream, err := os.OpenFile(filePath, os.O_RDONLY|os.O_SYNC, 0644)

	if os.IsNotExist(err) {
		return
	}

	if err != nil {
		panic(err)
	}

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
			v := reflect.New(t).Interface()
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
