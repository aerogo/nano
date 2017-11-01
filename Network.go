package nano

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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
			if networkSet(msg, db) == nil {
				serverBroadcast(client, msg)
			} else {
				fmt.Println("skip set broadcast")
			}

		case packetDelete:
			if networkDelete(msg, db) == nil {
				serverBroadcast(client, msg)
			} else {
				fmt.Println("skip delete broadcast")
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

			close(collection.loaded)

		case packetSet:
			// TODO: This can be optimized to use worker threads and channels
			go networkSet(msg, node)

		case packetDelete:
			// TODO: This can be optimized to use worker threads and channels
			go networkDelete(msg, node)
		}
	}
}

// serverBroadcast ...
func serverBroadcast(client *server.Client, msg *packet.Packet) {
	fromRemoteClient := client.Node.IsRemoteAddress(client.Connection.RemoteAddr())

	for targetClient := range client.Node.AllClients() {
		// Ignore the client who sent us the packet in the first place
		if targetClient == client {
			continue
		}

		// Do not send packets from remote clients to other remote clients.
		// Every node is responsible for notifying other remote nodes about changes.
		if fromRemoteClient && client.Node.IsRemoteAddress(targetClient.Connection.RemoteAddr()) {
			continue
		}

		// Send packet
		targetClient.Outgoing <- msg
	}
}

// networkSet ...
func networkSet(msg *packet.Packet, db *Node) error {
	data := bytes.NewBuffer(msg.Data)

	packetTimeBuffer := make([]byte, 8, 8)
	data.Read(packetTimeBuffer)
	packetTime, err := packet.Int64FromBytes(packetTimeBuffer)

	if err != nil {
		return err
	}

	namespaceName := readLine(data)
	namespace := db.Namespace(namespaceName)

	collectionName := readLine(data)
	collection := namespace.Collection(collectionName)
	<-collection.loaded

	key := readLine(data)

	jsonBytes, _ := data.ReadBytes('\n')
	jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("\n"))

	value := reflect.New(collection.typ).Interface()
	err = json.Unmarshal(jsonBytes, &value)

	if err != nil {
		return err
	}

	// Check timestamp
	lastModificationObj, exists := collection.lastModification.Load(key)

	if exists {
		lastModification := lastModificationObj.(int64)

		if packetTime < lastModification {
			return errors.New("Outdated packet")
		}
	}

	// Perform the actual set
	collection.set(key, value)

	// Update last modification time
	collection.lastModification.Store(key, packetTime)

	return nil
}

// networkDelete ...
func networkDelete(msg *packet.Packet, db *Node) error {
	data := bytes.NewBuffer(msg.Data)

	packetTimeBuffer := make([]byte, 8, 8)
	data.Read(packetTimeBuffer)
	packetTime, err := packet.Int64FromBytes(packetTimeBuffer)

	if err != nil {
		return err
	}

	namespaceName := readLine(data)
	namespace := db.Namespace(namespaceName)

	collectionName := readLine(data)
	collection := namespace.Collection(collectionName)
	<-collection.loaded

	key := readLine(data)

	// Check timestamp
	obj, exists := collection.lastModification.Load(key)

	if exists {
		lastModification := obj.(int64)

		if packetTime < lastModification {
			return errors.New("Outdated packet")
		}
	}

	// Perform the actual deletion
	collection.delete(key)

	// Update last modification time
	collection.lastModification.Store(key, packetTime)

	return nil
}

func readLine(data *bytes.Buffer) string {
	line, _ := data.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	return line
}
