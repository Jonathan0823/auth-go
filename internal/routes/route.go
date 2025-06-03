package routes

import (
	"github.com/Jonathan0823/auth-go/internal/handler"
	"github.com/Jonathan0823/auth-go/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, mainHandler *handler.MainHandler) {
	api := r.Group("/api")
	auth := api.Group("/auth")
	{
		provider := auth.Group("/:provider")
		provider.Use(middleware.OAuthMiddleware)
		{
			provider.GET("/", mainHandler.OAuthLogin)
			provider.GET("/callback", mainHandler.OAuthCallback)
		}
		auth.POST("/register", mainHandler.Register)
		auth.POST("/login", mainHandler.Login)
		auth.POST("/logout", mainHandler.Logout)
		auth.POST("/refresh", mainHandler.Refresh)
		auth.POST("/forgot-password", mainHandler.ForgotPassword)
		auth.POST("/reset-password", mainHandler.ResetPassword)
		verify := auth.Group("/verify")
		{
			verify.GET("/email", mainHandler.VerifyEmail)
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
