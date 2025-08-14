// Package middleware provides middleware functions for authentication and OAuth handling in a Gin web application.
package middleware

import (
	"log"
	"net/http"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.Error(errors.Unauthorized("Unauthorized: missing token", err))
			c.Abort()
			return
		}

		user, err := utils.ValidateJWT(token, "access")
		if err != nil {
			c.Error(errors.Unauthorized("Unauthorized: invalid token", err))
			c.Abort()
			return
		}

		c.Set("user", user)

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
			if appErr, ok := err.(*errors.Error); ok {
				if appErr.Err != nil {
					log.Println("Internal error:", appErr.Err)
				}
				c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			c.Abort()
		}
	}
}
