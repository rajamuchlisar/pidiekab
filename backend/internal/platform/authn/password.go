package authn

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash indicates that the encoded hash string is malformed.
	ErrInvalidHash = errors.New("invalid argon2id hash format")
	// ErrIncompatibleVersion indicates an unsupported argon2 version.
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Argon2idParams holds the memory, iterations, and parallelism parameters.
type Argon2idParams struct {
	Memory      uint32 // in KiB (e.g. 64 * 1024 = 65536 KiB = 64 MB)
	Iterations  uint32 // number of passes
	Parallelism uint8  // number of threads
	SaltLength  uint32 // bytes of salt
	KeyLength   uint32 // bytes of generated key
}

// DefaultArgon2idParams returns baseline OWASP-recommended parameters.
func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// FastArgon2idParamsForTesting returns lightweight parameters for rapid test execution.
func FastArgon2idParamsForTesting() Argon2idParams {
	return Argon2idParams{
		Memory:      1024, // 1 MB
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HashPassword hashes a plaintext password using Argon2id and returns an encoded string:
// $argon2id$v=19$m=65536,t=3,p=2$<salt-b64>$<hash-b64>
func HashPassword(password string, p Argon2idParams) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Iterations, p.Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword checks if a plaintext password matches an encoded Argon2id hash.
// It executes in constant time to prevent timing attacks.
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	// Format: ["", "argon2id", "v=19", "m=...,t=...,p=...", "salt", "hash"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, ErrIncompatibleVersion
	}

	var p Argon2idParams
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, ErrInvalidHash
	}
	p.Memory = memory
	p.Iterations = iterations
	p.Parallelism = parallelism

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHash
	}
	p.KeyLength = uint32(len(expectedHash))

	computedHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	if subtle.ConstantTimeCompare(expectedHash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}
