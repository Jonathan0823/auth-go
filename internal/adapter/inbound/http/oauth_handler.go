package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/platform"
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
	if !h.allowRateLimit(c, ipRateLimit("oauth_ip", c)) {
		return
	}
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
	if !h.allowRateLimit(c, ipRateLimit("oauth_ip", c)) {
		return
	}
	userData, err := h.Svc.OAuth.OAuthLogin(ctx, c.Writer, c.Request, c.Param("provider"))
	if err != nil {
		authErr := fmt.Errorf("oauth authentication failed: %w", domain.ErrUnauthenticated)
		h.auditFailure(c, platform.EventAuthOAuth, c.Param("provider"), authErr, 0)
		c.Error(authErr)
		return
	}

	h.auditSuccess(c, platform.EventAuthOAuth, c.Param("provider"), userData.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged in successfully",
		"user":    dto.UserResponseFromDomain(userData),
	})
}
