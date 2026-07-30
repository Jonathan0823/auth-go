package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/Jonathan0823/auth-go/internal/bootstrap"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func main() {
	if os.Getenv("ENVIRONMENT") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	cfg := platform.LoadConfig()
	bootstrap.Run(cfg)
}
