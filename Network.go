package nano

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aerogo/cluster/client"
	"github.com/aerogo/cluster/server"
	"github.com/aerogo/packet"
	jsoniter "github.com/json-iterator/go"
)

// serverReadPacketsFromClient reads packets from clients on the server side.
func serverReadPacketsFromClient(client *packet.Stream, node *Node) {
	for msg := range client.Incoming {
		switch msg.Type {
		// case packetPing:
		// 	fmt.Println("client", string(msg.Data))
		// 	client.Outgoing <- packet.New(packetPong, []byte("pong"))

		// case packetPong:
		// 	fmt.Println("client", string(msg.Data))

		case packetCollectionRequest:
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := node.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			if node.verbose {
				fmt.Println("COLLECTION REQUEST", client.Connection().RemoteAddr(), namespaceName+"."+collectionName)
			}

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

			if node.verbose {
				fmt.Println("COLLECTION REQUEST ANSWERED", client.Connection().RemoteAddr())
			}

		case packetSet:
			if networkSet(msg, node) == nil {
				serverForwardPacket(node.Server(), client, msg)
			}

		case packetDelete:
			if networkDelete(msg, node) == nil {
				serverForwardPacket(node.Server(), client, msg)
			}

		default:
			fmt.Printf("Error: Unknown network packet type %d of length %d\n", msg.Type, msg.Length)
		}
	}

	if node.verbose {
		fmt.Println("[server] Client disconnected", client.Connection().RemoteAddr())
	}
}

// clientReadPacketsFromServer reads packets from the server on the client side.
func clientReadPacketsFromServer(client *client.Node, node *Node) {
	for msg := range client.Stream.Incoming {
		switch msg.Type {
		case packetCollectionResponse:
			data := bytes.NewBuffer(msg.Data)

			namespaceName, _ := data.ReadString('\n')
			namespaceName = strings.TrimSuffix(namespaceName, "\n")

			namespace := node.Namespace(namespaceName)

			collectionName, _ := data.ReadString('\n')
			collectionName = strings.TrimSuffix(collectionName, "\n")

			if node.verbose {
				fmt.Println("COLLECTION RESPONSE RECEIVED", client.Address(), namespaceName+"."+collectionName)
			}

			collection := namespace.collectionLoading(collectionName)
			collection.readRecords(data)

			namespace.collectionsLoading.Delete(collectionName)
			close(collection.loaded)

		case packetSet, packetDelete:
			node.networkWorkerQueue <- msg

		case packetServerClose:
			if node.verbose {
				fmt.Println("[client] Server closed!", client.Address())
			}

			client.Close()

			if node.verbose {
				fmt.Println("[client] Reconnecting", client.Address())
			}

			client.Connect()

			if node.verbose {
				fmt.Println("[client] Reconnect finished!", client.Address())
			}

		default:
			fmt.Printf("Error: Unknown network packet type %d of length %d\n", msg.Type, msg.Length)
		}
	}

	close(node.networkWorkerQueue)

	if node.verbose {
		fmt.Println(client.Address(), "clientReadPacketsFromServer goroutine stopped")
	}
}

// clientNetworkWorker runs in a separate goroutine and handles the set & delete packets.
func clientNetworkWorker(node *Node) {
	for msg := range node.networkWorkerQueue {
		switch msg.Type {
		case packetSet:
			networkSet(msg, node)

		case packetDelete:
			networkDelete(msg, node)
		}
	}
}

// networkSet performs a set operation based on the information in the network packet.
func networkSet(msg *packet.Packet, db *Node) error {
	data := bytes.NewBuffer(msg.Data)

	packetTimeBuffer := make([]byte, 8)
	data.Read(packetTimeBuffer)
	packetTime, err := packet.Int64FromBytes(packetTimeBuffer)

	if err != nil {
		return err
	}

	namespaceName := readLine(data)
	namespace := db.Namespace(namespaceName)

	collectionName := readLine(data)
	collectionObj, exists := namespace.collections.Load(collectionName)

	if !exists || collectionObj == nil {
		return errors.New("Received networkSet command on non-existing collection")
	}

	collection := collectionObj.(*Collection)
	key := readLine(data)

	jsonBytes, _ := data.ReadBytes('\n')
	jsonBytes = bytes.TrimSuffix(jsonBytes, []byte("\n"))

	value := reflect.New(collection.typ).Interface()
	err = jsoniter.Unmarshal(jsonBytes, &value)

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

// networkDelete performs a delete operation based on the information in the network packet.
func networkDelete(msg *packet.Packet, db *Node) error {
	data := bytes.NewBuffer(msg.Data)

	packetTimeBuffer := make([]byte, 8)
	data.Read(packetTimeBuffer)
	packetTime, err := packet.Int64FromBytes(packetTimeBuffer)

	if err != nil {
		return err
	}

	namespaceName := readLine(data)
	namespace := db.Namespace(namespaceName)

	collectionName := readLine(data)
	collectionObj, exists := namespace.collections.Load(collectionName)

	if !exists || collectionObj == nil {
		return errors.New("Received networkDelete command on non-existing collection")
	}

	collection := collectionObj.(*Collection)
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

// serverOnConnect returns a function that can be used as a parameter
// for the OnConnect method. It is called every time a new client connects
// to the node.
func serverOnConnect(node *Node) func(*packet.Stream) {
	return func(stream *packet.Stream) {
		if node.verbose {
			fmt.Println("[server] New client", stream.Connection().RemoteAddr())
		}

		// Start reading packets from the client
		go serverReadPacketsFromClient(stream, node)
	}
}

// serverForwardPacket forwards the packet from the given client to other clients.
func serverForwardPacket(serverNode *server.Node, client *packet.Stream, msg *packet.Packet) {
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
		select {
		case targetClient.Outgoing <- msg:
			// Send successful.
		default:
			// Discard packet.
			// TODO: Find a better solution to deal with this.
		}
	}
}

// readLine reads a single line from the byte buffer and will not include the line break character.
func readLine(data *bytes.Buffer) string {
	line, _ := data.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	return line
}
