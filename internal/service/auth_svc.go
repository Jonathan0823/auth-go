// Package service provides business logic for user authentication and management
package service

import (
	"database/sql"
	goerror "errors"
	"fmt"
	"os"
	"time"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) Register(user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalServerError("failed to hash password", err)
	}

	userFromDB, err := s.repo.GetUserByEmail(user.Email, false)
	if userFromDB != nil && err == nil {
		return errors.Conflict("email already exists", err)
	}

	user.Password = string(hashedPassword)
	if err := s.repo.CreateUser(user); err != nil {
		return errors.InternalServerError("failed to create user", err)
	}

	if err := s.CreateVerifyEmail(user.Email); err != nil {
		return errors.InternalServerError("failed to create verification email", err)
	}

	return nil
}

func (s *service) Login(user models.User) (string, string, error) {
	userFromDB, err := s.repo.GetUserByEmail(user.Email, true)
	if err != nil && userFromDB == nil {
		if goerror.Is(err, sql.ErrNoRows) {
			return "", "", errors.NotFound("user not found", err)
		}
		return "", "", errors.InternalServerError("failed to get user by email", err)
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
		UserID:           userFromDB.ID,
		JTI:              jtiRefresh,
		RefreshedFromJTI: nil,
		InvalidatedAt:    nil,
		ExpiredAt:        time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:        time.Now(),
		IPAddress:        user.IPAddress,
		UserAgent:        user.UserAgent,
	}

	if err := s.repo.CreateTokenLog(tokenLog); err != nil {
		return "", "", errors.InternalServerError("failed to create token log", err)
	}

	return accessToken, refreshToken, nil
}

func (s *service) CreateVerifyEmail(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email, false)
	if err != nil || userFromDB == nil {
		if goerror.Is(err, sql.ErrNoRows) {
			return errors.NotFound("user not found", err)
		}
		return errors.InternalServerError("failed to get user by email", err)
	}

	verifyEmail := models.VerifyEmail{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.repo.CreateVerifyEmail(verifyEmail); err != nil {
		return errors.InternalServerError("failed to create verification email", err)
	}

	if err := utils.SendEmail(email, "Verify Email", "Click here to verify your email"); err != nil {
		return errors.InternalServerError("failed to send verification email", err)
	}

	return nil
}

func (s *service) VerifyEmail(id string, c *gin.Context) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return errors.BadRequest("invalid token", err)
	}

	verifyEmail, err := s.repo.GetVerifyEmailByID(id)
	if err != nil || verifyEmail.ID == uuid.Nil {
		return errors.InternalServerError("internal server error", err)
	}

	if time.Now().After(verifyEmail.ExpiredAt) {
		return errors.BadRequest("token expired", err)
	}

	if err := s.repo.VerifyEmail(id); err != nil {
		return errors.InternalServerError("failed to verify email", err)
	}
	return nil
}

func (s *service) ForgotPassword(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email, false)
	if err != nil || userFromDB == nil {
		if goerror.Is(err, sql.ErrNoRows) {
			return errors.NotFound("user not found", err)
		}
		return errors.InternalServerError("failed to get user by email", err)
	}

	data := models.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateForgotPasswordEmail(data); err != nil {
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

func (s *service) ResetPassword(id string, newPassword string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.BadRequest("invalid token", err)
	}

	forgotPassword, err := s.repo.GetForgotPasswordByID(id)
	if err != nil || forgotPassword.ID == uuid.Nil {
		return errors.InternalServerError("internal server error", err)
	}

	if time.Now().After(forgotPassword.ExpiredAt) {
		return errors.BadRequest("token expired", nil)
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalServerError("failed to hash new password", err)
	}

	if err = s.repo.UpdateUserPassword(forgotPassword.UserID, string(hashedNewPassword)); err != nil {
		return errors.InternalServerError("failed to update user password", err)
	}

	if err := s.repo.DeleteForgotPasswordByID(id); err != nil {
		return errors.InternalServerError("failed to delete forgot password record", err)
	}
	return nil
}

func (s *service) RefreshTokens(refreshToken, ip, userAgent string) (string, string, error) {
	claims, err := utils.ValidateJWT(refreshToken, "refresh")
	if err != nil {
		return "", "", errors.Unauthorized("invalid refresh token", err)
	}

	oldJTI := claims["jti"].(string)
	isJWTInvalidated, err := s.IsTokenLogInvalidated(oldJTI)
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

	if err := s.InvalidateJWTTokens(oldJTI, newJTI); err != nil {
		return "", "", errors.InternalServerError("failed to invalidate old tokens", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *service) InvalidateJWTTokens(oldJTI, newJTI string) error {
	if oldJTI == "" || newJTI == "" {
		return errors.BadRequest("oldJTI and newJTI cannot be empty", nil)
	}
	if err := s.repo.InvalidateTokenLog(oldJTI, newJTI); err != nil {
		return errors.InternalServerError("failed to invalidate token log", err)
	}
	return nil
}

func (s *service) IsTokenLogInvalidated(jti string) (bool, error) {
	if jti == "" {
		return false, errors.BadRequest("jti cannot be empty", nil)
	}
	invalidated, err := s.repo.IsTokenLogInvalidated(jti)
	if err != nil {
		return false, errors.InternalServerError("failed to check if token log is invalidated", err)
	}
	return invalidated, nil
}
