package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func main() {
	key, err := GenerateKey()
	if err != nil {
		panic(err)

	}

	fmt.Printf("Generated secret key: %s\n", hex.EncodeToString(key))
}
