package utils

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetUser(c *gin.Context) (dto.User, error) {
	claims, exists := c.Get("user")
	if !exists {
		return dto.User{}, fmt.Errorf("User is not found")
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return dto.User{}, fmt.Errorf("Invalid token claims")
	}

	return dto.User{
		ID:       int(mapClaims["id"].(float64)),
		Username: mapClaims["username"].(string),
		Email:    mapClaims["email"].(string),
	}, nil
}
