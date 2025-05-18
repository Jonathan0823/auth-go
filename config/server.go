package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewServer() *Config {
	return &Config{
		Port:           os.Getenv("PORT"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
	}
}

func (config *Config) InitServer(r *gin.Engine) {
	if config.Port == "" {
		config.Port = "8080"
	}

	if config.AllowedOrigins == "" || config.Port == "" {
		log.Fatal("ALLOWED_ORIGINS or PORT is not set in .env file")
	}

	r.Use(gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.AllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Run(fmt.Sprintf(":%s", config.Port))
}
