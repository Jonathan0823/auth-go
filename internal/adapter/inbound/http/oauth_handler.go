package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func (h *Handler) OAuthLogin(c *gin.Context) {
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *Handler) OAuthCallback(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.Error(fmt.Errorf("oauth authentication failed: %w", domain.ErrUnauthenticated))
		return
	}

	user := domain.User{
		OAuthID:   gothUser.UserID,
		Email:     gothUser.Email,
		Username:  gothUser.NickName,
		Provider:  gothUser.Provider,
		AvatarURL: gothUser.AvatarURL,
	}

	userData, err := h.Svc.OAuth.OAuthLogin(ctx, user)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user":    dto.UserResponseFromDomain(userData),
	})
}
