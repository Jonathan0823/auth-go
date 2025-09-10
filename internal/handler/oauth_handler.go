package handler

import (
	"net/http"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func (h *MainHandler) OAuthLogin(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *MainHandler) OAuthCallback(c *gin.Context) {
	ctx, cancel := utils.CtxWithTimeOut(c)
	defer cancel()
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.Error(errors.Unauthorized("OAuth authentication failed", err))
		return
	}

	data := models.User{
		OAuthID:   user.UserID,
		Email:     user.Email,
		Username:  user.NickName,
		Provider:  user.Provider,
		AvatarURL: user.AvatarURL,
	}

	userData, err := h.svc.OAuth().OAuthLogin(ctx, data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user":    userData,
	})
}
