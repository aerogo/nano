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

// serverReadPacketsFromClient ...
func serverReadPacketsFromClient(client *packet.Stream, node *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		// case packetPing:
		// 	fmt.Println("client", string(msg.Data))
		// 	client.Outgoing <- packet.New(packetPong, []byte("pong"))

		// case packetPong:
		// 	fmt.Println("client", string(msg.Data))

		case packetCollectionRequest:
			fmt.Println("COLLECTION REQUEST", client.Connection().RemoteAddr())
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := node.Namespace(namespaceName)

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
			fmt.Println("COLLECTION REQUEST ANSWERED", client.Connection().RemoteAddr())

		case packetSet:
			if networkSet(msg, node) == nil {
				serverBroadcast(node.Server(), client, msg)
			}

		case packetDelete:
			if networkDelete(msg, node) == nil {
				serverBroadcast(node.Server(), client, msg)
			}

		case packetClose:
			client.Close()

		default:
			fmt.Printf("Error: Unknown network packet type %d of length %d\n", msg.Type, msg.Length)
		}
	}
}

// clientReadPackets ...
func clientReadPackets(client *client.Node, node *Node) {
	for msg := range client.Stream.Incoming {
		switch msg.Type {
		// case packetPing:
		// 	fmt.Println("server", string(msg.Data))
		// 	client.Outgoing <- packet.New(packetPong, []byte("pong"))

		// case packetPong:
		// 	fmt.Println("server", string(msg.Data))

		case packetCollectionResponse:
			fmt.Println("COLLECTION RESPONSE RECEIVED", client.Address())
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := node.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			collection := namespace.collectionLoading(collectionName)
			collection.readRecords(data)

			namespace.collectionsLoading.Delete(collectionName)
			close(collection.loaded)

		case packetSet:
			node.networkSetQueue <- msg

		case packetDelete:
			node.networkDeleteQueue <- msg

		case packetClose:
			fmt.Println("[client] Server closed!", client.Address())
			client.Close()
			client.Connect()

		default:
			fmt.Printf("Error: Unknown network packet type %d of length %d\n", msg.Type, msg.Length)
		}
	}

	close(node.networkSetQueue)
	close(node.networkDeleteQueue)
	fmt.Println(client.Address(), "clientReadPackets goroutine stopped")
}

// clientNetworkWorker ...
func clientNetworkWorker(node *Node) {
	for {
		select {
		case msg, ok := <-node.networkSetQueue:
			if !ok {
				return
			}

			networkSet(msg, node)

		case msg, ok := <-node.networkDeleteQueue:
			if !ok {
				return
			}

			networkDelete(msg, node)
		}
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

// serverOnConnect ...
func serverOnConnect(db *Node) func(*packet.Stream) {
	return func(stream *packet.Stream) {
		fmt.Println("[server] New client", stream.Connection().RemoteAddr())

		// Start reading packets from the client
		go serverReadPacketsFromClient(stream, db)
	}
}

// serverBroadcast ...
func serverBroadcast(serverNode *server.Node, client *packet.Stream, msg *packet.Packet) {
	fromRemoteClient := serverNode.IsRemoteAddress(client.Connection().RemoteAddr())

	for targetClient := range serverNode.AllClients() {
		// Ignore the client who sent us the packet in the first place
		if targetClient == client {
			continue
		}

		// Do not send packets from remote clients to other remote clients.
		// Every node is responsible for notifying other remote nodes about changes.
		if fromRemoteClient && serverNode.IsRemoteAddress(targetClient.Connection().RemoteAddr()) {
			continue
		}

		// Send packet
		targetClient.Outgoing <- msg
	}
}

// readLine ...
func readLine(data *bytes.Buffer) string {
	line, _ := data.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	return line
}
