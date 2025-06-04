package handler

import (
	"net/http"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func (h *MainHandler) OAuthLogin(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *MainHandler) OAuthCallback(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := models.User{
		OAuthID:   user.UserID,
		Email:     user.Email,
		Username:  user.NickName,
		Provider:  user.Provider,
		AvatarURL: user.AvatarURL,
	}

	userData, err := h.svc.OAuthLogin(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user":    userData,
	})
}
