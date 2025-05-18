package handler

import (
	"net/http"
	"os"

	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/gin-gonic/gin"
)

func (h *MainHandler) Register(c *gin.Context) {
	var user dto.User
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
	var user dto.User
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

	c.SetCookie("access_token", access_token, 15*60, "/", domain, false, false)

	c.SetCookie("refresh_token", refresh_token, 7*24*3600, "/", domain, false, true)

	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully", "token": gin.H{"access_token": access_token, "refresh_token": refresh_token}})
}
