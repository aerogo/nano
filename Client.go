package nano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/aerogo/cluster/client"
	"github.com/aerogo/packet"
)

// clientReadPackets ...
func clientReadPackets(client *client.Node, db *Database) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPing:
			fmt.Println(string(msg.Data))
			client.Outgoing <- packet.New(packetPong, []byte("pong"))

		case packetCollection:
			data := bytes.NewBuffer(msg.Data)
			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := db.Collection(collectionName)
			collection.readRecords(data)

		case packetSet:
			onSet(msg, db)
		}
	}
}

// onSet ...
func onSet(msg *packet.Packet, db *Database) {
	data := bytes.NewBuffer(msg.Data)

	collectionName, _ := data.ReadString('\n')
	collectionName = strings.TrimSuffix(collectionName, "\n")
	collection := db.Collection(collectionName)

	key, _ := data.ReadString('\n')
	key = strings.TrimSuffix(key, "\n")

	jsonBytes, _ := data.ReadBytes('\n')
	jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("\n"))

	value := reflect.New(collection.typ).Interface()
	json.Unmarshal(jsonBytes, &value)

	collection.set(key, value)
}
