// Package handler provides HTTP handlers for user authentication and management
package handler

import (
	"net/http"
	"os"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	domain := os.Getenv("DOMAIN")
	c.SetCookie("access_token", "", -1, "/", domain, secure, false)
	c.SetCookie("refresh_token", "", -1, "/", domain, secure, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (h *MainHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}

	claims, err := utils.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user := models.User{
		Username:  claims["username"].(string),
		Email:     claims["email"].(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	newAccessToken, _, err := utils.GenerateJWT(user, "access")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	oldJTI := claims["jti"].(string)
	newRefreshToken, newJTI, err := utils.GenerateJWT(user, "refresh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	if err := h.svc.InvalidateJWTTokens(oldJTI, newJTI); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invalidate old tokens"})
		return
	}

	c.SetCookie("access_token", newAccessToken, 7*24*3600, "/", domain, secure, false)
	c.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

func (h *MainHandler) VerifyEmail(c *gin.Context) {
	id := c.Query("id")

	if err := h.svc.VerifyEmail(id, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *MainHandler) ResendVerifyEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	if err := h.svc.CreateVerifyEmail(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
