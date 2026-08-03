package jwt

import (
	"os"
	"testing"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func setupTokenEnv(t *testing.T) func() {
	origAccess := os.Getenv("JWT_ACCESS_SECRET")
	origHashKey := os.Getenv("REFRESH_TOKEN_HASH_KEY")
	_ = os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-32-chars-long-for-hs256!")
	_ = os.Setenv("REFRESH_TOKEN_HASH_KEY", "test-refresh-hash-key-32-chars-long-for-test")

	return func() {
		_ = os.Setenv("JWT_ACCESS_SECRET", origAccess)
		_ = os.Setenv("REFRESH_TOKEN_HASH_KEY", origHashKey)
	}
}

func TestGenerateAccessToken(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()
	user := domain.User{ID: 1, Username: "testuser", Email: "test@example.com"}

	token, jti, err := s.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}
	if jti == "" {
		t.Fatal("GenerateAccessToken() returned empty jti")
	}
}

func TestValidateAccessToken(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()
	user := domain.User{ID: 1, Username: "testuser", Email: "test@example.com"}

	token, _, err := s.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	claims, err := s.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims["id"] != float64(1) {
		t.Fatalf("ValidateAccessToken() id = %v, want 1", claims["id"])
	}
	if claims["username"] != "testuser" {
		t.Fatalf("ValidateAccessToken() username = %v, want testuser", claims["username"])
	}
	if claims["email"] != "test@example.com" {
		t.Fatalf("ValidateAccessToken() email = %v, want test@example.com", claims["email"])
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()
	_, err := s.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Fatal("ValidateAccessToken() with invalid token: expected error")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()

	raw1, hash1, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if raw1 == "" {
		t.Fatal("GenerateRefreshToken() returned empty raw token")
	}
	if len(hash1) == 0 {
		t.Fatal("GenerateRefreshToken() returned empty hash")
	}

	raw2, _, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() 2nd call error = %v", err)
	}

	if raw1 == raw2 {
		t.Fatal("Two refresh tokens should differ")
	}
}

func TestHashRefreshToken(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()

	raw, hash1, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	hash2, err := s.HashRefreshToken(raw)
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}

	if len(hash1) != 32 {
		t.Fatalf("expected 32-byte HMAC, got hash1=%d", len(hash1))
	}

	if len(hash2) != 32 {
		t.Fatalf("expected 32-byte HMAC, got hash2=%d", len(hash2))
	}

	for i := range hash1 {
		if hash1[i] != hash2[i] {
			t.Fatal("HashRefreshToken() result differs from GenerateRefreshToken() result")
		}
	}
}

func TestGenerateRefreshTokenNotJWT(t *testing.T) {
	cleanup := setupTokenEnv(t)
	defer cleanup()

	s := NewTokenService()
	raw, _, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// Refresh tokens must not contain JWT separators (dots).
	for _, c := range raw {
		if c == '.' {
			t.Fatal("Refresh token looks like a JWT (contains '.')")
		}
	}
}
