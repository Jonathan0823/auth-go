package main

import (
	"fmt"
	"log"

	"github.com/Jonathan0823/auth-go/config"
	"github.com/Jonathan0823/auth-go/internal/handler"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/Jonathan0823/auth-go/internal/routes"
	"github.com/Jonathan0823/auth-go/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config.InitOAuth()

	db := config.InitDB()
	defer db.Close()

	r := gin.New()
	r.Use(gin.Logger())
	repo := repository.NewRepository(db)
	svc := service.NewService(*repo)
	mainHandler := handler.NewMainHandler(svc)

	routes.RegisterRoutes(r, mainHandler)

	config.NewServer().InitServer(r)

	fmt.Println("Server is running on port 8080")
}
