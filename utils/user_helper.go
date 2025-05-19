package utils

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetUser(c *gin.Context) (models.User, error) {
	claims, exists := c.Get("user")
	if !exists {
		return models.User{}, fmt.Errorf("User is not found")
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return models.User{}, fmt.Errorf("Invalid token claims")
	}

	return models.User{
		ID:       int(mapClaims["id"].(float64)),
		Username: mapClaims["username"].(string),
		Email:    mapClaims["email"].(string),
	}, nil
}
