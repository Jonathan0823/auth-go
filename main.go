package main

import (
	"fmt"
	"log"

	"github.com/Jonathan0823/auth-go/config"
	"github.com/Jonathan0823/auth-go/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.InitDB()
	defer db.Close()

	r := gin.New()

	config.NewServer().InitServer(r)

	routes.RegisterRoutes(r, db)

	fmt.Println("Server is running on port 8080")
}
