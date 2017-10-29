package nano

import (
	"bytes"
	"encoding/binary"
)

// fromBytes converts a Big Endian representation to an int64
func fromBytes(b []byte) (int64, error) {
	buffer := bytes.NewReader(b)
	var result int64
	err := binary.Read(buffer, binary.BigEndian, &result)
	return result, err
}

// toBytes converts an int64 to its Big Endian representation
func toBytes(i int64) []byte {
	buffer := &bytes.Buffer{}
	binary.Write(buffer, binary.BigEndian, i)
	return buffer.Bytes()
}
