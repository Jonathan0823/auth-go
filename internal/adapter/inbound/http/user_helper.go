package http

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetUser(c *gin.Context) (domain.User, error) {
	claims, exists := c.Get("user")
	if !exists {
		return domain.User{}, fmt.Errorf("user is not found")
	}
	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return domain.User{}, fmt.Errorf("invalid token claims")
	}
	return domain.User{
		ID:       int(mapClaims["id"].(float64)),
		Username: mapClaims["username"].(string),
		Email:    mapClaims["email"].(string),
	}, nil
}
