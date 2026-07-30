package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

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
			c.Error(domain.Unauthorized("Unauthorized: missing token", err))
			c.Abort()
			return
		}
		claims, err := m.Tokens.ValidateToken(token, "access")
		if err != nil {
			c.Error(domain.Unauthorized("Unauthorized: invalid token", err))
			c.Abort()
			return
		}
		c.Set("user", claims)
		c.Next()
	}
}

func OAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		if provider == "" {
			provider = "github"
		}
		gothic.GetProviderName = func(req *http.Request) (string, error) {
			return provider, nil
		}
		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			if appErr, ok := err.(*domain.Error); ok {
				if appErr.Err != nil {
					log.Println("Internal error:", appErr.Err)
				}
				code, msg := http.StatusInternalServerError, "internal server error"
				switch appErr.Code {
				case domain.ErrCodeBadRequest:
					code = http.StatusBadRequest
					msg = appErr.Message
				case domain.ErrCodeNotFound:
					code = http.StatusNotFound
				case domain.ErrCodeConflict:
					code = http.StatusConflict
				case domain.ErrCodeUnauthorized:
					code = http.StatusUnauthorized
				case domain.ErrCodeForbidden:
					code = http.StatusForbidden
				}
				c.JSON(code, gin.H{"error": msg})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			c.Abort()
		}
	}
}
