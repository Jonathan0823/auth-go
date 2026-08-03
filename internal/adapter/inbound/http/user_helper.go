// Package http provides inbound HTTP handlers and routing.
package http

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func GetUser(c *gin.Context) (domain.User, error) {
	raw, exists := c.Get("user")
	if !exists {
		return domain.User{}, fmt.Errorf("user is not found")
	}
	claims, ok := raw.(map[string]any)
	if !ok {
		return domain.User{}, fmt.Errorf("invalid token claims")
	}
	return domain.User{
		ID:       int(claims["id"].(float64)),
		Username: claims["username"].(string),
		Email:    claims["email"].(string),
	}, nil
}
