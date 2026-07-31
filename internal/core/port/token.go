package port

import "github.com/Jonathan0823/auth-go/internal/core/domain"

type TokenService interface {
	GenerateAccessToken(user domain.User) (token string, jti string, err error)
	ValidateAccessToken(tokenString string) (map[string]interface{}, error)
	GenerateRefreshToken() (rawToken string, hmacHash []byte, err error)
	HashRefreshToken(rawToken string) ([]byte, error)
}
