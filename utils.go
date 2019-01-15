package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
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

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// CompareHash returns true if hashsrc == hashdst, false if hashsrc != hashdst
func CompareHash(hashsrc, hashdst []byte) bool {
	return string(hashsrc) == string(hashdst)
}

// IntToByte returns bytes of int
func IntToByte(num int) []byte {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		fmt.Printf("\rerr occur: %s\n", err)
		os.Exit(1)
	}
	return buffer.Bytes()
}

// gobEncode retuns
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// MapToSlice returns the map of the slice
func MapToSlice(oldMap map[string]string) []string {
	slice := []string{}

	for _, v := range oldMap {
		slice = append(slice, v)
	}

	return slice
}
