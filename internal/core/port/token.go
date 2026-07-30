package port

import "github.com/Jonathan0823/auth-go/internal/core/domain"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	RefreshJTI   string
}

type TokenService interface {
	GenerateAccessToken(user domain.User) (token, jti string, err error)
	GenerateRefreshToken(user domain.User) (token, jti string, err error)
	ValidateToken(tokenString, tokenType string) (map[string]interface{}, error)
}
