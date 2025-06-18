// Package middleware provides middleware functions for authentication and OAuth handling in a Gin web application.
package middleware

import (
	"net/http"

	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func AuthMiddleware(c *gin.Context) {
	token, err := c.Cookie("access_token")
	if err != nil || token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing token"})
		return
	}

	user, err := utils.ValidateJWT(token, "access")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.Set("user", user)

	c.Next()
}

func OAuthMiddleware(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		provider = "github"
	}

	gothic.GetProviderName = func(req *http.Request) (string, error) {
		return provider, nil
	}

	c.Next()
}
