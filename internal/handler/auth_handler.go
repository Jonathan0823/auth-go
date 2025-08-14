// Package handler provides HTTP handlers for user authentication and management
package handler

import (
	"net/http"
	"os"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

var secure = os.Getenv("ENVIRONMENT") == "production"

func (h *MainHandler) Register(c *gin.Context) {
	var user models.User
	if isValid := utils.BindJSONWithValidation(c, &user); !isValid {
		return
	}

	if err := h.svc.Register(user); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (h *MainHandler) Login(c *gin.Context) {
	var user models.User
	if isValid := utils.BindJSONWithValidation(c, &user); !isValid {
		return
	}

	user.IPAddress = c.ClientIP()
	user.UserAgent = c.GetHeader("User-Agent")

	accessToken, refreshToken, err := h.svc.Login(user)
	if err != nil {
		c.Error(err)
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	c.SetCookie("access_token", accessToken, 7*24*3600, "/", domain, secure, false)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", domain, secure, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (h *MainHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(errors.Unauthorized("Refresh token not found"))
		return
	}

	claims, err := utils.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		c.Error(errors.Unauthorized("Invalid refresh token"))
	}

	oldJTI := claims["jti"].(string)
	if err := h.svc.InvalidateJWTTokens(oldJTI, ""); err != nil {
		c.Error(err)
		return
	}
	domain := os.Getenv("DOMAIN")
	c.SetCookie("access_token", "", -1, "/", domain, secure, false)
	c.SetCookie("refresh_token", "", -1, "/", domain, secure, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (h *MainHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(errors.Unauthorized("Refresh token not found"))
		return
	}

	newAccessToken, newRefreshToken, err := h.svc.RefreshTokens(refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.Error(err)
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	c.SetCookie("access_token", newAccessToken, 7*24*3600, "/", domain, secure, false)
	c.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

func (h *MainHandler) VerifyEmail(c *gin.Context) {
	id := c.Query("id")

	if err := h.svc.VerifyEmail(id, c); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *MainHandler) ResendVerifyEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.Error(errors.BadRequest("Email is required"))
		return
	}

	if err := h.svc.CreateVerifyEmail(email); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email resent successfully"})
}

func (h *MainHandler) ForgotPassword(c *gin.Context) {
	type Request struct {
		Email string `json:"email" validate:"required,email"`
	}
	var user Request
	if isValid := utils.BindJSONWithValidation(c, &user); !isValid {
		return
	}

	if err := h.svc.ForgotPassword(user.Email); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent to your email"})
}

func (h *MainHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if isValid := utils.BindJSONWithValidation(c, &req); !isValid {
		return
	}

	if err := h.svc.ResetPassword(req.ID, req.Password); err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
