package nano

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/aerogo/cluster/client"
	"github.com/aerogo/cluster/server"
	"github.com/aerogo/packet"
)

// serverOnConnect ...
func serverOnConnect(db *Node) func(*server.Client) {
	return func(client *server.Client) {
		fmt.Println("New client", client.Connection.RemoteAddr())

		// Start reading packets from the client
		go serverReadPacketsFromClient(client, db)

		// // Send initial packet
		// client.Outgoing <- packet.New(messagePing, []byte("ping"))

		// // Send collection data
		// wg := sync.WaitGroup{}

		// for typeName := range db.types {
		// 	wg.Add(1)

		// 	go func(name string) {
		// 		collection := db.Collection(name)

		// 		var b bytes.Buffer
		// 		b.WriteString(collection.name)
		// 		b.WriteByte('\n')

		// 		writer := bufio.NewWriter(&b)
		// 		collection.writeRecords(writer, false)
		// 		writer.Flush()

		// 		client.Outgoing <- packet.New(packetCollection, b.Bytes())

		// 		wg.Done()
		// 	}(typeName)
		// }

		// wg.Wait()
	}
}

// serverReadPacketsFromClient ...
func serverReadPacketsFromClient(client *server.Client, db *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPong:
			fmt.Println(string(msg.Data))

		case packetSet:
			set(msg, db)

			for targetClient := range db.server.AllClients() {
				if targetClient == client {
					continue
				}

				targetClient.Outgoing <- msg
			}
		}
	}
}

// clientReadPackets ...
func clientReadPackets(client *client.Node, db *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPing:
			fmt.Println(string(msg.Data))
			client.Outgoing <- packet.New(packetPong, []byte("pong"))

		case packetCollectionResponse:
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := db.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := namespace.Collection(collectionName)
			collection.readRecords(data)

		case packetSet:
			set(msg, db)
		}
	}
}

// set ...
func set(msg *packet.Packet, db *Node) {
	data := bytes.NewBuffer(msg.Data)

	namespaceName, _ := data.ReadString('\n')
	namespaceName = strings.TrimSuffix(namespaceName, "\n")
	namespace := db.Namespace(namespaceName)

	collectionName, _ := data.ReadString('\n')
	collectionName = strings.TrimSuffix(collectionName, "\n")
	collection := namespace.Collection(collectionName)

	key, _ := data.ReadString('\n')
	key = strings.TrimSuffix(key, "\n")

	jsonBytes, _ := data.ReadBytes('\n')
	jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("\n"))

	value := reflect.New(collection.typ).Interface()
	json.Unmarshal(jsonBytes, &value)

	collection.set(key, value)
}
