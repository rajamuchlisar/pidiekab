package authn

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	// SessionTokenBytes is the number of random bytes in a raw session token (256 bits).
	SessionTokenBytes = 32
)

// GenerateRandomToken generates a cryptographically secure hex-encoded random string.
func GenerateRandomToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", errors.New("byte length must be greater than zero")
	}

	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to read secure random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// GenerateSessionToken creates a 256-bit secure session token (64 hex characters).
func GenerateSessionToken() (string, error) {
	return GenerateRandomToken(SessionTokenBytes)
}

// HashToken computes a SHA-256 hash of a raw token string and returns it as lowercase hex.
// This is stored at rest in the database instead of the raw token.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
