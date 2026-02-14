package nanoid

import (
	"crypto/rand"
	"fmt"
)

// alphabet is URL-safe: A-Za-z0-9_-
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

// DefaultLength is the default NanoID length (12 chars = 72 bits of entropy)
const DefaultLength = 12

// Generate creates a cryptographically secure NanoID of the given length.
// Uses a 64-char URL-safe alphabet (A-Za-z0-9_-) with crypto/rand.
func Generate(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("nanoid: failed to read random bytes: %w", err)
	}

	id := make([]byte, length)
	for i := 0; i < length; i++ {
		// Mask to 6 bits (0-63) to index into 64-char alphabet with no bias
		id[i] = alphabet[bytes[i]&63]
	}
	return string(id), nil
}

// New creates a NanoID with the default length (12 chars).
func New() (string, error) {
	return Generate(DefaultLength)
}

// MustNew creates a NanoID with the default length, panicking on error.
func MustNew() string {
	id, err := New()
	if err != nil {
		panic(err)
	}
	return id
}
