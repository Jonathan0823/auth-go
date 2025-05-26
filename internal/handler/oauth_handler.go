package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func (h *MainHandler) OAuthCallback(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *MainHandler) OAuthLogin(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	fmt.Println("User logged in:", user)

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
	})
}
