package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters:
//   - time=3, memory=64MB, threads=4 — tuned for ~200ms on modern hardware
//   - 16-byte random salt, 32-byte derived key
//   - Benchmark and adjust for your production environment.
const (
	argon2idTime    = 3
	argon2idMemory  = 64 * 1024
	argon2idThreads = 4
	argon2idKeyLen  = 32
	argon2idSaltLen = 16
)

type hasher struct{}

func NewHasher() *hasher { return &hasher{} }

func (h *hasher) Hash(password string) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)

	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	keyB64 := base64.RawStdEncoding.EncodeToString(key)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2idMemory, argon2idTime, argon2idThreads, saltB64, keyB64)
	return encoded, nil
}

func (h *hasher) Compare(hashedPassword, plainPassword string) error {
	params, salt, key, err := decodeArgon2id(hashedPassword)
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	otherKey := argon2.IDKey([]byte(plainPassword), salt, params.time, params.memory, params.threads, params.keyLen)

	if subtle.ConstantTimeCompare(key, otherKey) != 1 {
		return fmt.Errorf("password mismatch")
	}
	return nil
}

type argon2idParams struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

func decodeArgon2id(encoded string) (*argon2idParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid encoded hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unexpected algorithm: %s", parts[1])
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, fmt.Errorf("parse version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("unexpected version: %d", version)
	}

	params := &argon2idParams{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.threads); err != nil {
		return nil, nil, nil, fmt.Errorf("parse params: %w", err)
	}
	params.keyLen = argon2idKeyLen

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode salt: %w", err)
	}
	if len(salt) != argon2idSaltLen {
		return nil, nil, nil, fmt.Errorf("unexpected salt length: %d", len(salt))
	}

	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode key: %w", err)
	}
	if len(key) != argon2idKeyLen {
		return nil, nil, nil, fmt.Errorf("unexpected key length: %d", len(key))
	}

	return params, salt, key, nil
}
