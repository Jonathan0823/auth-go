package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Jonathan0823/auth-go/config"
	"github.com/Jonathan0823/auth-go/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db := config.InitDB()
	defer db.Close()

	r := gin.New()
	r.Use(gin.Logger())

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not set in .env file")
	}

	routes.InitRoutes(r, db)

	r.Run(fmt.Sprintf(":%s", port))
	fmt.Println("Server is running on port 8080")
}
