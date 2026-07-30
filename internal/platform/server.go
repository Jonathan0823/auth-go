package platform

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitServer(r *gin.Engine, cfg Config) {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.AllowedOrigins == "" {
		log.Fatal("ALLOWED_ORIGINS is not set")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Run(fmt.Sprintf(":%s", cfg.Port))
}
