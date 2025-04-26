package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// EncodeToBase64 encodes a string to base64
func EncodeToBase64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// DecodeFromBase64 decodes a base64 string
func DecodeFromBase64(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GenerateSHA256Hash returns the SHA256 hash of the input string
func GenerateSHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// GenerateRandomString generates a random string of the given length
func GenerateRandomString(n int64) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:n], nil
}

// EncryptAES encrypts data using AES-256-GCM
func EncryptAES(plainText string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key length must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptAES decrypts AES-256-GCM encrypted base64 string
func DecryptAES(cipherTextBase64 string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", errors.New("key length must be 32 bytes for AES-256")
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
