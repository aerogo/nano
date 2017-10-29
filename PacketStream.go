package nano

import (
	"fmt"
	"net"
)

// PacketStream ...
type PacketStream struct {
	incoming   chan *Packet
	outgoing   chan *Packet
	connection *net.TCPConn
}

// read ...
func (stream *PacketStream) read() {
	typeBuffer := make([]byte, 1)
	lengthBuffer := make([]byte, 8)

	for {
		_, err := stream.connection.Read(typeBuffer)

		if err != nil {
			fmt.Println("R Packet Type fail", err)
			break
		}

		_, err = stream.connection.Read(lengthBuffer)

		if err != nil {
			fmt.Println("R Packet Length fail", stream.connection.RemoteAddr(), err)
			break
		}

		length, err := fromBytes(lengthBuffer)

		if err != nil {
			fmt.Println("R Packet Length decode fail", stream.connection.RemoteAddr(), err)
			break
		}

		data := make([]byte, length)
		readLength := 0

		for readLength < len(data) {
			n, err := stream.connection.Read(data[readLength:])
			readLength += n

			if err != nil {
				fmt.Println("R Data read fail", stream.connection.RemoteAddr(), err)
				break
			}
		}

		if readLength < len(data) {
			fmt.Println("R Data read length fail", stream.connection.RemoteAddr(), err)
			break
		}

		// msg, err := ioutil.ReadAll(stream.connection)

		// if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		// 	fmt.Println("R Timeout", stream.connection.RemoteAddr())
		// 	break
		// }

		// if err != nil && err != io.EOF && strings.Contains(err.Error(), "connection reset") {
		// 	fmt.Println("R Disconnected", stream.connection.RemoteAddr())
		// 	break
		// }

		stream.incoming <- NewPacket(typeBuffer[0], data)
	}
}

// write ...
func (stream *PacketStream) write() {
	for packet := range stream.outgoing {
		msg := packet.Bytes()
		totalWritten := 0

		for totalWritten < len(msg) {
			writtenThisCall, err := stream.connection.Write(msg[totalWritten:])

			if err != nil {
				fmt.Println("Error writing", err)
				return
			}

			totalWritten += writtenThisCall
		}
	}
}
