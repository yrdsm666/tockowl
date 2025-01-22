package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
)

// Uint32ToBytes convert uint32 to bytes
func Uint32ToBytes(n uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, n)
	return bytebuf.Bytes()
}

// BytesToUint32 convert bytes to uint32
func BytesToUint32(byt []byte) uint32 {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

// BytesToInt convert bytes to int
func BytesToInt(byt []byte) int {
	bytebuff := bytes.NewBuffer(byt)
	var data uint32
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

// IntToBytes convert int to bytes
func IntToBytes(n int) []byte {
	data := uint32(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

// Uint32sToBytes convert uint32s to bytes
func Uint32sToBytes(ns []uint32) []byte {
	bytebuf := bytes.NewBuffer([]byte{})
	for _, n := range ns {
		binary.Write(bytebuf, binary.BigEndian, n)
	}
	return bytebuf.Bytes()
}

// BytesToUint32s convert bytes to uint32s
func BytesToUint32s(byt []byte) []uint32 {
	bytebuff := bytes.NewBuffer(byt)
	data := make([]uint32, len(byt)/4)
	binary.Read(bytebuff, binary.BigEndian, &data)
	return data
}

func GetTxs(batchSize int, payloadSize int) []byte {
	reader := io.NopCloser(rand.Reader)

	txs := make([][]byte, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := make([]byte, payloadSize)
		n, err := reader.Read(tx)
		if err != nil && err != io.EOF {
			// if we get an error other than EOF
			log.Printf("An error occurred: %v", err)
			return nil
		} else if err == io.EOF && n == 0 {
			log.Println("Reached end of file. Sending empty commands until last command is executed...")
			return nil
		}
		txs[i] = tx
	}

	res, err := json.Marshal(txs)
	if err != nil {
		log.Printf("Marshal txs error: %v", err)
		return nil
	}

	return res
}
