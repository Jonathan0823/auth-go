package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type AuthMiddleware struct {
	Tokens port.TokenService
}

func NewAuthMiddleware(tokens port.TokenService) *AuthMiddleware {
	return &AuthMiddleware{Tokens: tokens}
}

func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			_ = c.Error(fmt.Errorf("missing access token: %w", domain.ErrUnauthenticated))
			c.Abort()
			return
		}
		claims, err := m.Tokens.ValidateAccessToken(token)
		if err != nil {
			_ = c.Error(fmt.Errorf("invalid access token: %w", domain.ErrUnauthenticated))
			c.Abort()
			return
		}
		c.Set("user", claims)
		c.Next()
	}
}
