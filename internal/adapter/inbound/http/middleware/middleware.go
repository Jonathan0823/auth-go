package middleware

import (
	"fmt"
	"log/slog"
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
			c.Error(fmt.Errorf("missing access token: %w", domain.ErrUnauthenticated))
			c.Abort()
			return
		}
		claims, err := m.Tokens.ValidateToken(token, "access")
		if err != nil {
			c.Error(fmt.Errorf("invalid access token: %w", domain.ErrUnauthenticated))
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

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status, message := mapError(err)
		if statusIsServerError(status) {
			logger.ErrorContext(c.Request.Context(), "application error",
				slog.String("request_id", RequestIDFromContext(c.Request.Context())),
				slog.Int("status", status),
				slog.String("error", message),
				slog.String("error_type", fmt.Sprintf("%T", err)),
			)
		}
		c.JSON(status, gin.H{"error": message})
		c.Abort()
	}
}
