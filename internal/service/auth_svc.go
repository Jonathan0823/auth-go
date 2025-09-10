// Package service provides business logic for user authentication and management
package service

import (
	"context"
	"database/sql"
	goerror "errors"
	"fmt"
	"os"
	"time"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, user models.User) error
	Login(ctx context.Context, user models.User) (string, string, error)
	ForgotPassword(ctx context.Context, email string) error
	CreateVerifyEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, id string, c *gin.Context) error
	ResetPassword(ctx context.Context, tokenStr string, newPassword string) error
	RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (string, string, error)
	InvalidateJWTTokens(ctx context.Context, oldJTI, newJTI string) error
	IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error)
}

type authService struct {
	repo repository.Repository
}

func NewAuthService(repo repository.Repository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) Register(ctx context.Context, user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalServerError("failed to hash password", err)
	}

	user.Password = string(hashedPassword)
	if err := s.repo.Users().CreateUser(ctx, user); err != nil {
		if utils.IsPGUniqueViolation(err) {
			return errors.Conflict("email already exists", err)
		}
		return errors.InternalServerError("failed to create user", err)
	}

	if err := s.CreateVerifyEmail(ctx, user.Email); err != nil {
		return errors.InternalServerError("failed to create verification email", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, user models.User) (string, string, error) {
	userFromDB, err := s.repo.Users().GetUserByEmail(ctx, user.Email, true)
	if err != nil {
		return "", "", errors.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return "", "", errors.NotFound("user not found", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password)); err != nil {
		return "", "", errors.Unauthorized("invalid credentials", err)
	}

	accessToken, _, err := utils.GenerateJWT(*userFromDB, "access")
	if err != nil {
		return "", "", errors.InternalServerError("failed to generate access token", err)
	}

	refreshToken, jtiRefresh, err := utils.GenerateJWT(*userFromDB, "refresh")
	if err != nil {
		return "", "", errors.InternalServerError("failed to generate refresh token", err)
	}

	tokenLog := models.TokenLog{
		ID:               uuid.New(),
		UserID:           userFromDB.ID,
		JTI:              jtiRefresh,
		RefreshedFromJTI: nil,
		InvalidatedAt:    nil,
		ExpiredAt:        time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:        time.Now(),
		IPAddress:        user.IPAddress,
		UserAgent:        user.UserAgent,
	}

	if err := s.repo.Auth().CreateTokenLog(ctx, tokenLog); err != nil {
		return "", "", errors.InternalServerError("failed to create token log", err)
	}

	return accessToken, refreshToken, nil
}

func (s *authService) CreateVerifyEmail(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetUserByEmail(ctx, email, false)
	if err != nil {
		return errors.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return errors.NotFound("user not found", nil)
	}

	verifyEmail := models.VerifyEmail{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.repo.Auth().CreateVerifyEmail(ctx, verifyEmail); err != nil {
		return errors.InternalServerError("failed to create verification email", err)
	}

	if err := utils.SendEmail(email, "Verify Email", "Click here to verify your email"); err != nil {
		return errors.InternalServerError("failed to send verification email", err)
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, id string, c *gin.Context) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.BadRequest("invalid token", err)
	}

	verifyEmail, err := s.repo.Auth().GetVerifyEmailByID(ctx, id)
	if err != nil || verifyEmail.ID == uuid.Nil {
		if goerror.Is(err, sql.ErrNoRows) {
			return errors.NotFound("verification token not found", err)
		}
		return errors.InternalServerError("internal server error", err)
	}

	if time.Now().After(verifyEmail.ExpiredAt) {
		return errors.BadRequest("token expired", err)
	}

	if err := s.repo.Auth().VerifyEmail(ctx, id); err != nil {
		return errors.InternalServerError("failed to verify email", err)
	}
	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetUserByEmail(ctx, email, false)
	if err != nil {
		return errors.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return errors.NotFound("user not found", nil)
	}

	data := models.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.Auth().CreateForgotPasswordEmail(ctx, data); err != nil {
		return errors.InternalServerError("failed to create forgot password record", err)
	}

	baseURL := os.Getenv("BASE_URL")
	if err := utils.SendEmail(email, "Password Reset", fmt.Sprintf(`
      Click here to reset your password: <a href="%s/reset-password?id=%s">Reset Password</a>`,
		baseURL, data.ID.String(),
	)); err != nil {
		return errors.InternalServerError("failed to send password reset email", err)
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, id string, newPassword string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.BadRequest("invalid token", err)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalServerError("failed to hash new password", err)
	}

	return s.repo.WithTx(ctx, func(u repository.UOW) error {
		forgotPassword, err := u.Auth().GetForgotPasswordByID(ctx, id)
		if err != nil || forgotPassword.ID == uuid.Nil {
			if goerror.Is(err, sql.ErrNoRows) {
				return errors.NotFound("forgot password token not found", err)
			}
			return errors.InternalServerError("internal server error", err)
		}

		if time.Now().After(forgotPassword.ExpiredAt) {
			return errors.BadRequest("token expired", nil)
		}

		if err := u.Auth().DeleteForgotPasswordByID(ctx, id); err != nil {
			return errors.InternalServerError("failed to delete forgot password record", err)
		}

		if err = u.Users().UpdateUserPassword(ctx, forgotPassword.UserID, string(hashedNewPassword)); err != nil {
			return errors.InternalServerError("failed to update user password", err)
		}
		return nil
	})
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (string, string, error) {
	claims, err := utils.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		return "", "", errors.Unauthorized("invalid refresh token", err)
	}

	oldJTI := claims["jti"].(string)
	isJWTInvalidated, err := s.IsTokenLogInvalidated(ctx, oldJTI)
	if err != nil && isJWTInvalidated {
		return "", "", errors.Unauthorized("invalidated refresh token", err)
	}

	user := models.User{
		Username:  claims["username"].(string),
		Email:     claims["email"].(string),
		IPAddress: ip,
		UserAgent: userAgent,
	}

	newAccessToken, _, err := utils.GenerateJWT(user, "access")
	if err != nil {
		return "", "", errors.InternalServerError("failed to generate access token", err)
	}

	newRefreshToken, newJTI, err := utils.GenerateJWT(user, "refresh")
	if err != nil {
		return "", "", errors.InternalServerError("failed to generate refresh token", err)
	}

	if err := s.InvalidateJWTTokens(ctx, oldJTI, newJTI); err != nil {
		return "", "", errors.InternalServerError("failed to invalidate old tokens", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) InvalidateJWTTokens(ctx context.Context, oldJTI, newJTI string) error {
	if oldJTI == "" || newJTI == "" {
		return errors.BadRequest("oldJTI and newJTI cannot be empty", nil)
	}
	if err := s.repo.Auth().InvalidateTokenLog(ctx, oldJTI, newJTI); err != nil {
		return errors.InternalServerError("failed to invalidate token log", err)
	}
	return nil
}

func (s *authService) IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, errors.BadRequest("jti cannot be empty", nil)
	}
	invalidated, err := s.repo.Auth().IsTokenLogInvalidated(ctx, jti)
	if err != nil {
		return false, errors.InternalServerError("failed to check if token log is invalidated", err)
	}
	return invalidated, nil
}
