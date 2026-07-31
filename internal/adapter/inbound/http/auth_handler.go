package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

var secure = os.Getenv("ENVIRONMENT") == "production"

const (
	accessCookieMaxAge  = 15 * 60
	refreshCookieMaxAge = 7 * 24 * 3600
)

func setAccessCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		MaxAge:   accessCookieMaxAge,
		Path:     "/",
		Domain:   cookieDomain(),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		MaxAge:   refreshCookieMaxAge,
		Path:     "/",
		Domain:   cookieDomain(),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookies(c *gin.Context) {
	clearCookie(c, "access_token")
	clearCookie(c, "refresh_token")
}

func clearCookie(c *gin.Context, name string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   cookieDomain(),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func cookieDomain() string {
	if d := os.Getenv("DOMAIN"); d != "" {
		return d
	}
	return "localhost"
}

func (h *Handler) Register(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.CredentialsRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}

	user := domain.User{Email: req.Email, Password: req.Password}
	if err := h.Svc.Auth.Register(ctx, user); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.CredentialsRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}

	user := domain.User{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}
	accessToken, refreshToken, err := h.Svc.Auth.Login(ctx, user)
	if err != nil {
		c.Error(err)
		return
	}

	setAccessCookie(c, accessToken)
	setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (h *Handler) Logout(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(fmt.Errorf("refresh token not found: %w", domain.ErrUnauthenticated))
		return
	}

	if err := h.Svc.Auth.Logout(ctx, refreshToken); err != nil {
		c.Error(err)
		return
	}

	clearCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(fmt.Errorf("refresh token not found: %w", domain.ErrUnauthenticated))
		return
	}

	newAccess, newRefresh, err := h.Svc.Auth.RefreshTokens(ctx, refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.Error(err)
		return
	}

	setAccessCookie(c, newAccess)
	setRefreshCookie(c, newRefresh)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	if err := h.Svc.Auth.VerifyEmail(ctx, c.Query("id")); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *Handler) ResendVerifyEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	if err := h.Svc.Auth.CreateVerifyEmail(ctx, email); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Verification email resent successfully"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.ForgotPasswordRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}
	if err := h.Svc.Auth.ForgotPassword(ctx, req.Email); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset link sent to your email"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req dto.ResetPasswordRequest
	if !BindJSONWithValidation(c, &req) {
		return
	}
	if err := h.Svc.Auth.ResetPassword(ctx, req.ID, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
