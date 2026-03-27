// Package crypto provides AES-GCM authenticated encryption utilities.
// Use EncryptionAESKey (16, 24, or 32 bytes) to select AES-128, AES-192, or AES-256.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plaintext with AES-GCM and returns a base64-encoded ciphertext.
// The nonce is prepended to the ciphertext and included in the base64 output.
func Encrypt(key []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decodes a base64-encoded ciphertext produced by Encrypt and returns
// the original plaintext.
func Decrypt(key []byte, encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptString is a convenience wrapper for string input/output.
func EncryptString(key []byte, plaintext string) (string, error) {
	return Encrypt(key, []byte(plaintext))
}

// DecryptString is a convenience wrapper that returns the plaintext as a string.
func DecryptString(key []byte, encoded string) (string, error) {
	b, err := Decrypt(key, encoded)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
