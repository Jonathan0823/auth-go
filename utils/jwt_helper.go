package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(user dto.User, jwtType string) (string, error) {
	var secretKey []byte
	if jwtType == "access" {
		secretKey = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	} else if jwtType == "refresh" {
		secretKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}

	if len(secretKey) == 0 {
		log.Fatal("JWT secret key is not set in the environment variables")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString string, jwtType string) (jwt.MapClaims, error) {
	var secretKey []byte
	if jwtType == "access" {
		secretKey = []byte(os.Getenv("JWT_ACCESS_SECRET"))
	} else if jwtType == "refresh" {
		secretKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	}

	if len(secretKey) == 0 {
		log.Fatal("JWT secret key is not set in the environment variables")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return *claims, nil
}
