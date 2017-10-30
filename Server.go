package nano

import (
	"bufio"
	"bytes"
	"fmt"
	"sync"

	"github.com/aerogo/cluster/server"
	"github.com/aerogo/packet"
)

// onConnect ...
func onConnect(db *Database) func(*server.Client) {
	return func(client *server.Client) {
		fmt.Println("New client", client.Connection.RemoteAddr())

		// Start reading packets from the client
		go serverReadPacketsFromClient(client, db)

		// // Send initial packet
		// client.Outgoing <- packet.New(messagePing, []byte("ping"))

		// Send collection data
		wg := sync.WaitGroup{}

		for typeName := range db.types {
			wg.Add(1)

			go func(name string) {
				collection := db.Collection(name)

				var b bytes.Buffer
				b.WriteString(collection.name)
				b.WriteByte('\n')

				writer := bufio.NewWriter(&b)
				collection.writeRecords(writer, false)
				writer.Flush()

				client.Outgoing <- packet.New(packetCollection, b.Bytes())

				wg.Done()
			}(typeName)
		}

		wg.Wait()
	}
}

// serverReadPacketsFromClient ...
func serverReadPacketsFromClient(client *server.Client, db *Database) {
	for msg := range client.Incoming {
		switch msg.Type {
		case packetPong:
			fmt.Println(string(msg.Data))

		case packetSet:
			onSet(msg, db)

			for obj := range db.server.AllClients() {
				targetClient := obj.(*server.Client)

				if targetClient == client {
					continue
				}

				targetClient.Outgoing <- msg
			}
		}
	}
}
