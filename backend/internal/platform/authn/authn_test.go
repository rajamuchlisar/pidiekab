package authn_test

import (
	"strings"
	"testing"

	"pidiekab/internal/platform/authn"
)

func TestHashPasswordAndVerify(t *testing.T) {
	params := authn.FastArgon2idParamsForTesting()
	password := "CorrectHorseBatteryStaple!2026"

	encoded, err := authn.HashPassword(password, params)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Errorf("unexpected hash prefix: %s", encoded)
	}

	// Verify with correct password
	match, err := authn.VerifyPassword(password, encoded)
	if err != nil {
		t.Fatalf("verification returned unexpected error: %v", err)
	}
	if !match {
		t.Error("expected password to match, got false")
	}

	// Verify with wrong password
	matchWrong, err := authn.VerifyPassword("WrongPassword123", encoded)
	if err != nil {
		t.Fatalf("verification with wrong password failed: %v", err)
	}
	if matchWrong {
		t.Error("expected wrong password not to match, got true")
	}

	// Empty password hashing should fail
	if _, err := authn.HashPassword("", params); err == nil {
		t.Error("expected error hashing empty password, got nil")
	}

	// Malformed hash strings should fail cleanly
	if _, err := authn.VerifyPassword(password, "not-a-valid-argon2-hash"); err == nil {
		t.Error("expected error verifying malformed hash, got nil")
	}

	if _, err := authn.VerifyPassword(password, "$bcrypt$v=19$bad$salt$hash"); err == nil {
		t.Error("expected error for non-argon2 hash, got nil")
	}
}

func TestGenerateAndHashToken(t *testing.T) {
	token1, err := authn.GenerateSessionToken()
	if err != nil {
		t.Fatalf("failed to generate session token 1: %v", err)
	}

	token2, err := authn.GenerateSessionToken()
	if err != nil {
		t.Fatalf("failed to generate session token 2: %v", err)
	}

	if len(token1) != 64 { // 32 bytes hex encoded = 64 characters
		t.Errorf("expected 64 hex characters, got %d", len(token1))
	}

	if token1 == token2 {
		t.Error("generated two identical tokens, expected cryptographically random unique tokens")
	}

	hash1 := authn.HashToken(token1)
	hash2 := authn.HashToken(token1) // same token
	hash3 := authn.HashToken(token2)

	if hash1 != hash2 {
		t.Errorf("hashing same token produced different outputs: %s vs %s", hash1, hash2)
	}

	if hash1 == hash3 {
		t.Errorf("hashing different tokens produced identical hash: %s", hash1)
	}

	if len(hash1) != 64 {
		t.Errorf("expected sha256 hex string length 64, got %d", len(hash1))
	}
}
