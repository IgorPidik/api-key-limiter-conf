package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"os"
)

func GenerateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func DecryptData(data string) (string, error) {
	secretKeyHex := os.Getenv("SECRET_KEY")
	secretKey, decodeErr := hex.DecodeString(secretKeyHex)
	if decodeErr != nil {
		return "", decodeErr
	}
	return Decrypt(secretKey, data)
}

func EncryptData(data string) (string, error) {
	secretKeyHex := os.Getenv("SECRET_KEY")
	secretKey, decodeErr := hex.DecodeString(secretKeyHex)
	if decodeErr != nil {
		return "", decodeErr
	}
	return Encrypt(secretKey, data)
}

func Encrypt(key []byte, data string) (string, error) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return hex.EncodeToString(ciphertext), nil
}

func Decrypt(key []byte, hexData string) (string, error) {
	data, decodeErr := hex.DecodeString(hexData)
	if decodeErr != nil {
		return "", nil
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return "", err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
