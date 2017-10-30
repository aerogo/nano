package nano

func onConnect() {
	// // Send initial packet
	// // client.Outgoing <- packet.New(messagePing, []byte("ping"))

	// // Send collection data
	// wg := sync.WaitGroup{}

	// for typeName := range server.db.types {
	// 	wg.Add(1)

	// 	go func(name string) {
	// 		collection := server.db.Collection(name)

	// 		var b bytes.Buffer
	// 		b.WriteString(collection.name)
	// 		b.WriteByte('\n')

	// 		writer := bufio.NewWriter(&b)
	// 		collection.writeRecords(writer, false)
	// 		writer.Flush()

	// 		client.Outgoing <- packet.New(messageCollection, b.Bytes())

	// 		wg.Done()
	// 	}(typeName)
	// }

	// wg.Wait()
}
