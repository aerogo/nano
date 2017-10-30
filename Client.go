package nano

// // readPackets ...
// func (client *Client) readPackets() {
// 	for msg := range client.Incoming {
// 		switch msg.Type {
// 		case messagePing:
// 			fmt.Println(string(msg.Data))
// 			client.Outgoing <- packet.New(messagePong, []byte("pong"))

// 		case messageCollection:
// 			data := bytes.NewBuffer(msg.Data)
// 			collectionName, _ := data.ReadString('\n')
// 			collectionName = strings.TrimSuffix(collectionName, "\n")

// 			collection := client.db.Collection(collectionName)
// 			collection.readRecords(data)

// 		case messageSet:
// 			onSet(msg, client.db)
// 		}
// 	}
// }
