package routes

import (
	"database/sql"

	"github.com/Jonathan0823/auth-go/internal/handler"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/Jonathan0823/auth-go/internal/service"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, db *sql.DB) {
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	mainHandler := handler.NewMainHandler(svc)

	api := r.Group("/api")
	auth := api.Group("/auth")
	{
		auth.POST("/register", mainHandler.Register)
		auth.POST("/login", mainHandler.Login)
	}

	user := api.Group("/user")
	{
		user.GET("/:id", mainHandler.GetUserByID)
	}

}
