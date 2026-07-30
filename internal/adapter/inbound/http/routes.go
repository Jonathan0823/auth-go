package http

import (
	"log/slog"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handler, logger *slog.Logger) {
	authMW := inhttp.NewAuthMiddleware(h.Tokens)

	api := r.Group("/api")
	api.Use(inhttp.ErrorHandler(logger))
	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		verify := auth.Group("/verify")
		{
			verify.GET("/email", h.VerifyEmail)
			verify.POST("/email/resend", h.ResendVerifyEmail)
		}
	}
	oauth := api.Group("/oauth")
	{
		provider := oauth.Group("/:provider")
		provider.Use(inhttp.OAuthMiddleware())
		{
			provider.GET("/", h.OAuthLogin)
			provider.GET("/callback", h.OAuthCallback)
		}
	}
	user := api.Group("/user")
	user.Use(authMW.Handler())
	{
		user.GET("/me", h.GetCurrentUser)
		user.GET("/:id", h.GetUserByID)
		user.GET("/get-all", h.GetAllUsers)
		user.GET("/email", h.GetUserByEmail)
		user.PATCH("/update", h.UpdateUser)
		user.DELETE("/delete/:id", h.DeleteUser)
	}
}
