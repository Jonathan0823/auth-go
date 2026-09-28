package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/observability"
)

const (
	accessCookieMaxAge  = 15 * 60
	refreshCookieMaxAge = 7 * 24 * 3600
)

func (h *Handler) setAccessCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		MaxAge:   accessCookieMaxAge,
		Path:     "/",
		Domain:   h.CookieDomain,
		Secure:   h.SecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		MaxAge:   refreshCookieMaxAge,
		Path:     "/",
		Domain:   h.CookieDomain,
		Secure:   h.SecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearCookies(c *gin.Context) {
	h.clearCookie(c, "access_token")
	h.clearCookie(c, "refresh_token")
}

func (h *Handler) clearCookie(c *gin.Context, name string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   h.CookieDomain,
		Secure:   h.SecureCookie,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Register creates a new user account and sends an email verification link.
// @Summary Register a user
// @ID register
// @Description Creates a user with an Argon2id-hashed password.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.CredentialsRequest true "Registration credentials"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.CredentialsRequest
	if !BindJSONWithValidation(c, &req) {
		h.audit(c, observability.EventUserRegister, observability.OutcomeFailure, observability.ReasonValidation, "", 0)
		return
	}

	if !h.allowRateLimit(c, ipRateLimit("register_ip", c)) {
		return
	}
	command := port.RegisterCommand{Email: req.Email, Password: req.Password}
	if err := h.Svc.Auth.Register(ctx, command); err != nil {
		h.auditFailure(c, observability.EventUserRegister, "", err, 0)
		_ = c.Error(err)
		return
	}
	h.auditSuccess(c, observability.EventUserRegister, "", 0)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "User registered successfully"})
}

// Login authenticates a user and sets access and refresh token cookies.
// @Summary Log in
// @ID login
// @Description Authenticates with email and password and sets HttpOnly access_token and refresh_token cookies.
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.CredentialsRequest true "Login credentials"
// @Success 200 {object} dto.MessageResponse
// @Header 200 {string} Set-Cookie "access_token=<jwt>; HttpOnly; SameSite=Lax"
// @Header 200 {string} Set-Cookie "refresh_token=<opaque-token>; HttpOnly; SameSite=Lax"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.CredentialsRequest
	if !BindJSONWithValidation(c, &req) {
		h.audit(c, observability.EventAuthLogin, observability.OutcomeFailure, observability.ReasonValidation, "", 0)
		return
	}

	if !h.allowRateLimit(c, ipRateLimit("login_ip", c)) {
		return
	}
	command := port.LoginCommand{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	accessToken, refreshToken, err := h.Svc.Auth.Login(ctx, command)
	if err != nil {
		h.auditFailure(c, observability.EventAuthLogin, "", err, 0)
		_ = c.Error(err)
		return
	}

	h.auditSuccess(c, observability.EventAuthLogin, "", 0)
	h.setAccessCookie(c, accessToken)
	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "User logged in successfully"})
}

// Logout revokes the refresh-token family and clears authentication cookies.
// @Summary Log out
// @ID logout
// @Tags authentication
// @Produce json
// @Security CookieAuth
// @Success 200 {object} dto.MessageResponse
// @Header 200 {string} Set-Cookie "access_token=; Max-Age=0"
// @Header 200 {string} Set-Cookie "refresh_token=; Max-Age=0"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		authErr := fmt.Errorf("refresh token not found: %w", domain.ErrUnauthenticated)
		h.auditFailure(c, observability.EventAuthLogout, "", authErr, 0)
		_ = c.Error(authErr)
		return
	}

	if err := h.Svc.Auth.Logout(ctx, refreshToken); err != nil {
		h.auditFailure(c, observability.EventAuthLogout, "", err, 0)
		_ = c.Error(err)
		return
	}

	h.auditSuccess(c, observability.EventAuthLogout, "", 0)
	h.clearCookies(c)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "User logged out successfully"})
}

// Refresh rotates the refresh token and sets a new access-token cookie.
// @Summary Refresh tokens
// @ID refreshTokens
// @Description Consumes the refresh_token cookie once and rotates the refresh-token family.
// @Tags authentication
// @Produce json
// @Security CookieAuth
// @Success 200 {object} dto.MessageResponse
// @Header 200 {string} Set-Cookie "access_token=<jwt>; HttpOnly; SameSite=Lax"
// @Header 200 {string} Set-Cookie "refresh_token=<opaque-token>; HttpOnly; SameSite=Lax"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	if !h.allowRateLimit(c, ipRateLimit("refresh_ip", c)) {
		return
	}
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		authErr := fmt.Errorf("refresh token not found: %w", domain.ErrUnauthenticated)
		h.auditFailure(c, observability.EventAuthRefresh, "", authErr, 0)
		_ = c.Error(authErr)
		return
	}

	newAccess, newRefresh, err := h.Svc.Auth.RefreshTokens(ctx, refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if isRefreshReplay(err) {
			h.audit(c, observability.EventAuthRefreshReplay, observability.OutcomeDetected, observability.ReasonReplayDetected, "", 0)
		} else {
			h.auditFailure(c, observability.EventAuthRefresh, "", err, 0)
		}
		_ = c.Error(err)
		return
	}

	h.auditSuccess(c, observability.EventAuthRefresh, "", 0)
	h.setAccessCookie(c, newAccess)
	h.setRefreshCookie(c, newRefresh)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Access token refreshed successfully"})
}

// VerifyEmail verifies an email address using the token in the query string.
// @Summary Verify email
// @ID verifyEmail
// @Tags authentication
// @Produce json
// @Param id query string true "Email verification token"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/auth/verify/email [get]
func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	if err := h.Svc.Auth.VerifyEmail(ctx, c.Query("id")); err != nil {
		h.auditFailure(c, observability.EventAuthEmailVerification, "", err, 0)
		_ = c.Error(err)
		return
	}
	h.auditSuccess(c, observability.EventAuthEmailVerification, "", 0)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Email verified successfully"})
}

// ResendVerifyEmail sends a new email verification link.
// @Summary Resend email verification
// @ID resendVerifyEmail
// @Tags authentication
// @Produce json
// @Param email query string true "Email address"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/auth/verify/email/resend [post]
func (h *Handler) ResendVerifyEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	email := c.Query("email")
	if email == "" {
		h.audit(c, observability.EventAuthEmailVerification, observability.OutcomeFailure, observability.ReasonValidation, "", 0)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if !h.allowRateLimit(c, ipRateLimit("verify_ip", c)) {
		return
	}
	if err := h.Svc.Auth.CreateVerifyEmail(ctx, email); err != nil {
		h.auditFailure(c, observability.EventAuthEmailVerification, "", err, 0)
		_ = c.Error(err)
		return
	}
	h.auditSuccess(c, observability.EventAuthEmailVerification, "", 0)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Verification email resent successfully"})
}

// ForgotPassword sends a password-reset link to an existing email address.
// @Summary Request password reset
// @ID forgotPassword
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email address"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.ForgotPasswordRequest
	if !BindJSONWithValidation(c, &req) {
		h.audit(c, observability.EventAuthPasswordResetRequest, observability.OutcomeFailure, observability.ReasonValidation, "", 0)
		return
	}
	if !h.allowRateLimit(c, ipRateLimit("recovery_ip", c)) {
		return
	}
	if err := h.Svc.Auth.ForgotPassword(ctx, req.Email); err != nil {
		h.auditFailure(c, observability.EventAuthPasswordResetRequest, "", err, 0)
		_ = c.Error(err)
		return
	}
	h.auditSuccess(c, observability.EventAuthPasswordResetRequest, "", 0)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password reset link sent to your email"})
}

// ResetPassword replaces a user's password with a valid reset token.
// @Summary Reset password
// @ID resetPassword
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.ResetPasswordRequest
	if !BindJSONWithValidation(c, &req) {
		h.audit(c, observability.EventAuthPasswordReset, observability.OutcomeFailure, observability.ReasonValidation, "", 0)
		return
	}
	if !h.allowRateLimit(c, ipRateLimit("recovery_ip", c)) {
		return
	}
	if err := h.Svc.Auth.ResetPassword(ctx, req.ID, req.Password); err != nil {
		h.auditFailure(c, observability.EventAuthPasswordReset, "", err, 0)
		_ = c.Error(err)
		return
	}
	h.auditSuccess(c, observability.EventAuthPasswordReset, "", 0)
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Password reset successfully"})
}
