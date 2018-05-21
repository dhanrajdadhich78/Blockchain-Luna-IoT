package blockchain

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func DbExists(dbFile string) (bool, error) {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false, err
	}

	return true, nil
}
