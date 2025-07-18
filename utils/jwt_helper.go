// Package utils is a utility package that provides functions for program's utility operations
package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateJWT(user models.User, jwtType string) (jwtToken string, jti string, err error) {
	var secretKey []byte
	switch jwtType {
	case "access":
		secretKey = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	case "refresh":
		secretKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}

	var expirationTime time.Time
	switch jwtType {
	case "access":
		expirationTime = time.Now().Add(time.Minute * 15)
	case "refresh":
		expirationTime = time.Now().Add(time.Hour * 24 * 7)
	}

	if len(secretKey) == 0 {
		log.Fatal("JWT secret key is not set in the environment variables")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	id := uuid.New().String()

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["jti"] = id
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["exp"] = expirationTime.Unix()

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, id, nil
}

func ValidateJWT(tokenString string, jwtType string) (jwt.MapClaims, error) {
	var secretKey []byte
	switch jwtType {
	case "access":
		secretKey = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	case "refresh":
		secretKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}

	if len(secretKey) == 0 {
		log.Fatal("JWT secret key is not set in the environment variables")
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
