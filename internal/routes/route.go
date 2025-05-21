package routes

import (
	"database/sql"

	"github.com/Jonathan0823/auth-go/internal/handler"
	"github.com/Jonathan0823/auth-go/internal/middleware"
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
		auth.POST("/logout", mainHandler.Logout)
		auth.POST("/refresh", mainHandler.Refresh)
		auth.POST("/forgot-password", mainHandler.ForgotPassword)
		verify := auth.Group("/verify")
		{
			verify.POST("/email", mainHandler.VerifyEmail)
			verify.POST("/email/resend", mainHandler.ResendVerifyEmail)
		}
	}

	user := api.Group("/user")
	user.Use(middleware.AuthMiddleware)
	{
		user.GET("/:id", mainHandler.GetUserByID)
		user.GET("/get-all", mainHandler.GetAllUsers)
		user.GET("/email", mainHandler.GetUserByEmail)
		user.PATCH("/update", mainHandler.UpdateUser)
		user.DELETE("/delete/:id", mainHandler.DeleteUser)
	}

}
