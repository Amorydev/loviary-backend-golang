package shared

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// NewUUID generates a new UUID
func NewUUID() string {
	return uuid.New().String()
}

// RandomToken generates a random token of specified byte length
func RandomToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// SecureRandomBytes generates secure random bytes
func SecureRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// Contains checks if a string is in a slice
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
