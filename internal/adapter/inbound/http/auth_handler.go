package http

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

var secure = os.Getenv("ENVIRONMENT") == "production"

func (h *Handler) Register(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	var req domain.LoginRegisterRequest
	if isValid := BindJSONWithValidation(c, &req); !isValid {
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
	var req domain.LoginRegisterRequest
	if isValid := BindJSONWithValidation(c, &req); !isValid {
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

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}
	c.SetCookie("access_token", accessToken, 7*24*3600, "/", domain, secure, false)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "User logged in successfully"})
}

func (h *Handler) Logout(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(domain.Unauthorized("Refresh token not found", err))
		return
	}
	claims, err := jwt.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		c.Error(domain.Unauthorized("Invalid refresh token", err))
		return
	}
	oldJTI := claims["jti"].(string)
	if err := h.Svc.Auth.InvalidateJWTTokens(ctx, oldJTI, ""); err != nil {
		c.Error(err)
		return
	}
	domain := os.Getenv("DOMAIN")
	c.SetCookie("access_token", "", -1, "/", domain, secure, false)
	c.SetCookie("refresh_token", "", -1, "/", domain, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func (h *Handler) Refresh(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.Error(domain.Unauthorized("Refresh token not found", err))
		return
	}

	newAccess, newRefresh, err := h.Svc.Auth.RefreshTokens(ctx, refreshToken, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		c.Error(err)
		return
	}

	d := os.Getenv("DOMAIN")
	if d == "" {
		d = "localhost"
	}
	c.SetCookie("access_token", newAccess, 7*24*3600, "/", d, secure, false)
	c.SetCookie("refresh_token", newRefresh, 7*24*3600, "/", d, secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Access token refreshed successfully"})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	ctx, cancel := CtxWithTimeout(c)
	defer cancel()
	id := c.Query("id")
	if err := h.Svc.Auth.VerifyEmail(ctx, id); err != nil {
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
		c.Error(domain.BadRequest("Email is required", nil))
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
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if isValid := BindJSONWithValidation(c, &req); !isValid {
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
	var req domain.ResetPasswordRequest
	if isValid := BindJSONWithValidation(c, &req); !isValid {
		return
	}
	if err := h.Svc.Auth.ResetPassword(ctx, req.ID, req.Password); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
