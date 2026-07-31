package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

// OAuthLogin starts the OAuth authorization flow for a supported provider.
// @Summary Start OAuth login
// @ID oauthLogin
// @Tags oauth
// @Param provider path string true "OAuth provider" Enums(github,google)
// @Success 302
// @Failure 400 {object} ErrorResponse
// @Router /api/oauth/{provider}/ [get]
func (h *Handler) OAuthLogin(c *gin.Context) {
	h.Svc.OAuth.BeginAuth(c.Writer, c.Request, c.Param("provider"))
}

// OAuthCallback completes the OAuth authorization flow and returns the authenticated user.
// @Summary Complete OAuth login
// @ID oauthCallback
// @Tags oauth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(github,google)
// @Success 200 {object} UserResponseEnvelope
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/oauth/{provider}/callback [get]
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
