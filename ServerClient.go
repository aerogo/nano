package nano

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/aerogo/packet"
)

// // readPackets ...
// func readPackets(client *server.Client) {
// 	for msg := range client.Incoming {
// 		switch msg.Type {
// 		case messagePong:
// 			fmt.Println(string(msg.Data))

// 		case messageSet:
// 			onSet(msg, client.server.db)

// 			for obj := range client.server.AllConnections() {
// 				targetClient := obj.(*ServerConnection)

// 				if targetClient == client {
// 					continue
// 				}

// 				targetClient.Outgoing <- msg
// 			}
// 		}
// 	}
// }

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
