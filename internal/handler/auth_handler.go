package handler

import (
	"net/http"
	"os"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (h *MainHandler) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
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
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	access_token, refresh_token, err := h.svc.Login(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	c.SetCookie("access_token", access_token, 7*24*3600, "/", domain, false, false)

	c.SetCookie("refresh_token", refresh_token, 7*24*3600, "/", domain, false, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (h *MainHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", os.Getenv("DOMAIN"), false, false)
	c.SetCookie("refresh_token", "", -1, "/", os.Getenv("DOMAIN"), false, true)

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

	accessToken, err := utils.GenerateJWT(models.User{Username: claims["username"].(string), Email: claims["email"].(string)}, "access")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	c.SetCookie("access_token", accessToken, 7*24*3600, "/", domain, false, false)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

func (h *MainHandler) VerifyEmail(c *gin.Context) {
	tokenStr := c.Query("token")

	if err := h.svc.VerifyEmail(tokenStr, c); err != nil {
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
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.svc.ForgotPassword(user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent to your email"})
}

func (h *MainHandler) ResetPassword(c *gin.Context) {
	type ResetPasswordRequest struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.svc.ResetPassword(req.Token, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
