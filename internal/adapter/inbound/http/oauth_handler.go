package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func (h *Handler) OAuthLogin(c *gin.Context) {
	h.Svc.OAuth.BeginAuth(c.Writer, c.Request, c.Param("provider"))
}

func (h *Handler) OAuthCallback(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	userData, err := h.Svc.OAuth.OAuthLogin(ctx, c.Writer, c.Request, c.Param("provider"))
	if err != nil {
		c.Error(fmt.Errorf("oauth authentication failed: %w", domain.ErrUnauthenticated))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user":    dto.UserResponseFromDomain(userData),
	})
}
