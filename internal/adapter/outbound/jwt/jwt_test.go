package jwt

import (
	"testing"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

const (
	testAccessSecret = "test-access-secret-32-chars-long-for-hs256!"
	testRefreshKey   = "test-refresh-hash-key-32-chars-long-for-test"
)

func newTestTokenService() *tokenService {
	return NewTokenService(testAccessSecret, testRefreshKey).(*tokenService)
}

func TestGenerateAccessToken(t *testing.T) {
	s := newTestTokenService()
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
	s := newTestTokenService()
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

func TestValidateAccessTokenInvalid(t *testing.T) {
	s := newTestTokenService()
	if _, err := s.ValidateAccessToken("invalid-token"); err == nil {
		t.Fatal("ValidateAccessToken() with invalid token: expected error")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	s := newTestTokenService()
	raw1, hash1, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if raw1 == "" || len(hash1) == 0 {
		t.Fatal("GenerateRefreshToken() returned empty token or hash")
	}

	raw2, _, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() second call error = %v", err)
	}
	if raw1 == raw2 {
		t.Fatal("two refresh tokens should differ")
	}
}

func TestHashRefreshToken(t *testing.T) {
	s := newTestTokenService()
	raw, hash1, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	hash2, err := s.HashRefreshToken(raw)
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	if len(hash1) != 32 || len(hash2) != 32 {
		t.Fatalf("expected 32-byte HMAC, got %d and %d", len(hash1), len(hash2))
	}
	if string(hash1) != string(hash2) {
		t.Fatal("HashRefreshToken() result differs from GenerateRefreshToken() result")
	}
}

func TestGenerateRefreshTokenNotJWT(t *testing.T) {
	s := newTestTokenService()
	raw, _, err := s.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	for _, c := range raw {
		if c == '.' {
			t.Fatal("refresh token looks like a JWT")
		}
	}
}

func TestEmptySecretsReturnErrors(t *testing.T) {
	s := NewTokenService("", "")
	if _, _, err := s.GenerateAccessToken(domain.User{}); err == nil {
		t.Fatal("GenerateAccessToken succeeded without a secret")
	}
	if _, _, err := s.GenerateRefreshToken(); err == nil {
		t.Fatal("GenerateRefreshToken succeeded without a key")
	}
}
