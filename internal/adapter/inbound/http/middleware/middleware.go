package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

	"github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {
			c.Error(domain.Unauthorized("Unauthorized: missing token", err))
			c.Abort()
			return
		}
		claims, err := jwt.ValidateJWT(token, "access")
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
				c.JSON(appErr.Code, gin.H{"error": appErr.Message})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
			c.Abort()
		}
	}
}
