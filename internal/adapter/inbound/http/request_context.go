package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthenticatedUser struct {
	ID       int
	Username string
	Email    string
}

func CtxWithTimeout(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 10*time.Second)
}

func GetUser(c *gin.Context) (AuthenticatedUser, error) {
	raw, exists := c.Get("user")
	if !exists {
		return AuthenticatedUser{}, fmt.Errorf("user is not found")
	}
	claims, ok := raw.(map[string]any)
	if !ok {
		return AuthenticatedUser{}, fmt.Errorf("invalid token claims")
	}
	id, ok := claims["id"].(float64)
	if !ok || id <= 0 || id != float64(int(id)) {
		return AuthenticatedUser{}, fmt.Errorf("invalid token user ID")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return AuthenticatedUser{}, fmt.Errorf("invalid token username")
	}
	email, ok := claims["email"].(string)
	if !ok {
		return AuthenticatedUser{}, fmt.Errorf("invalid token email")
	}
	return AuthenticatedUser{ID: int(id), Username: username, Email: email}, nil
}
