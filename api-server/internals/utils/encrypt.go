package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

func newGCM() (cipher.AEAD, error) {
	encodedKey := os.Getenv("ENCRYPTION_KEY")

	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func Encrypt(plainText string) ([]byte, error) {
	gcm, err := newGCM()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plainText), nil)

	return append(nonce, ciphertext...), nil
}

func Decrypt(encryptedValue []byte) (string, error) {
	gcm, err := newGCM()
	if err != nil {
		return "", err
	}
	if len(encryptedValue) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce := encryptedValue[:gcm.NonceSize()]
	ciphertext := encryptedValue[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
