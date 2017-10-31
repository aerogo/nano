package nano

import (
	"bufio"
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
	}
}

// serverReadPacketsFromClient ...
func serverReadPacketsFromClient(client *server.Client, db *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPong:
			fmt.Println(string(msg.Data))

		case packetCollectionRequest:
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := db.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := namespace.Collection(collectionName)

			var b bytes.Buffer

			b.WriteString(namespace.name)
			b.WriteByte('\n')

			b.WriteString(collection.name)
			b.WriteByte('\n')

			writer := bufio.NewWriter(&b)
			collection.writeRecords(writer, false)
			writer.Flush()

			client.Outgoing <- packet.New(packetCollectionResponse, b.Bytes())

		case packetSet:
			set(msg, db)

			fromRemoteClient := db.Server().IsRemoteAddress(client.Connection.RemoteAddr())
			fmt.Println("from remote", client.Connection.RemoteAddr(), fromRemoteClient)

			for targetClient := range db.server.AllClients() {
				if targetClient == client {
					continue
				}

				if fromRemoteClient && db.Server().IsRemoteAddress(targetClient.Connection.RemoteAddr()) {
					fmt.Println("skip remote", targetClient.Connection.RemoteAddr())
					continue
				}

				targetClient.Outgoing <- msg
			}
		}
	}
}

// clientReadPackets ...
func clientReadPackets(client *client.Node, node *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPing:
			fmt.Println(string(msg.Data))
			client.Outgoing <- packet.New(packetPong, []byte("pong"))

		case packetCollectionResponse:
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := node.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := namespace.Collection(collectionName)
			collection.readRecords(data)

			collection.loaded <- true

		case packetSet:
			go set(msg, node)
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
	err := json.Unmarshal(jsonBytes, &value)

	if err != nil {
		panic(err)
	}

	collection.set(key, value)
}
