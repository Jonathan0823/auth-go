package middleware

import (
	"net/http"

	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
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
