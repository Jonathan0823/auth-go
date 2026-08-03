package jwt

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type tokenService struct{}

func NewTokenService() port.TokenService {
	return &tokenService{}
}

func (s *tokenService) GenerateAccessToken(user domain.User) (token, jti string, err error) {
	secret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	if len(secret) == 0 {
		log.Fatal("JWT_ACCESS_SECRET is not set")
	}
	jti = uuid.New().String()
	claims := jwt.MapClaims{
		"id":       user.ID,
		"jti":      jti,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	return t, jti, err
}

func (s *tokenService) ValidateAccessToken(tokenString string) (map[string]any, error) {
	secret := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	if len(secret) == 0 {
		log.Fatal("JWT_ACCESS_SECRET is not set")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *tokenService) GenerateRefreshToken() (rawToken string, hmacHash []byte, err error) {
	key := []byte(os.Getenv("REFRESH_TOKEN_HASH_KEY"))
	if len(key) == 0 {
		log.Fatal("REFRESH_TOKEN_HASH_KEY is not set")
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, fmt.Errorf("generate random: %w", err)
	}

	rawToken = base64.RawURLEncoding.EncodeToString(bytes)

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(rawToken))
	hmacHash = mac.Sum(nil)

	return rawToken, hmacHash, nil
}

func (s *tokenService) HashRefreshToken(rawToken string) ([]byte, error) {
	key := []byte(os.Getenv("REFRESH_TOKEN_HASH_KEY"))
	if len(key) == 0 {
		return nil, fmt.Errorf("REFRESH_TOKEN_HASH_KEY is not set")
	}

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(rawToken))
	return mac.Sum(nil), nil
}
