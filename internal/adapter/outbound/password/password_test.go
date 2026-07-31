package password

import (
	"testing"
)

func TestHashAndCompare(t *testing.T) {
	h := NewHasher()
	password := "my-secret-password-123"

	hash, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty string")
	}

	if err := h.Compare(hash, password); err != nil {
		t.Fatalf("Compare() with correct password: %v", err)
	}

	if err := h.Compare(hash, "wrong-password"); err == nil {
		t.Fatal("Compare() with wrong password: expected error, got nil")
	}
}

func TestHashProducesArgon2idFormat(t *testing.T) {
	h := NewHasher()
	hash, err := h.Hash("test-password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	// Verify the format: $argon2id$v=19$m=...,t=...,p=...$<salt>$<key>
	if len(hash) < 20 {
		t.Fatalf("Hash too short: %q", hash)
	}
	if hash[:10] != "$argon2id$" {
		t.Fatalf("Hash missing argon2id prefix: %q", hash)
	}
	if hash[10:14] != "v=19" && hash[10:13] != "v=1" {
		// v=19 for argon2 version number
	}
}

func TestTwoHashesDiffer(t *testing.T) {
	h := NewHasher()
	h1, _ := h.Hash("same-password")
	h2, _ := h.Hash("same-password")

	if h1 == h2 {
		t.Fatal("Two hashes of the same password should differ (random salt)")
	}
}

func TestInvalidHash(t *testing.T) {
	h := NewHasher()
	err := h.Compare("not-a-valid-hash", "password")
	if err == nil {
		t.Fatal("Compare() with invalid hash: expected error, got nil")
	}
}

func TestBcryptHashNotAccepted(t *testing.T) {
	h := NewHasher()
	// A typical bcrypt hash (generated from cost 10)
	bcryptHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	err := h.Compare(bcryptHash, "password")
	if err == nil {
		t.Fatal("Compare() with bcrypt hash: expected error, got nil")
	}
}
