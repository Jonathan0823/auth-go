package jwt

import (
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

func (s *tokenService) GenerateRefreshToken(user domain.User) (token, jti string, err error) {
	secret := []byte(os.Getenv("JWT_REFRESH_SECRET"))
	if len(secret) == 0 {
		log.Fatal("JWT_REFRESH_SECRET is not set")
	}
	jti = uuid.New().String()
	claims := jwt.MapClaims{
		"id":       user.ID,
		"jti":      jti,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	return t, jti, err
}

func (s *tokenService) ValidateToken(tokenString, tokenType string) (map[string]interface{}, error) {
	var secret []byte
	switch tokenType {
	case "access":
		secret = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	case "refresh":
		secret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}
	if len(secret) == 0 {
		log.Fatal("JWT secret key is not set")
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

// ValidateJWT is a standalone helper used by the HTTP middleware.
func ValidateJWT(tokenString, tokenType string) (jwt.MapClaims, error) {
	var secret []byte
	switch tokenType {
	case "access":
		secret = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	case "refresh":
		secret = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}
	if len(secret) == 0 {
		log.Fatal("JWT secret key is not set")
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
