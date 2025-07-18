// Package service provides business logic for user authentication and management
package service

import (
	"fmt"
	"os"
	"time"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) Register(user models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userFromDB, err := s.repo.GetUserByEmail(user.Email, false)
	if userFromDB.Email != "" && err == nil {
		return fmt.Errorf("user with this email already exists")
	}

	user.Password = string(hashedPassword)
	if err := s.repo.CreateUser(user); err != nil {
		return err
	}

	if err := s.CreateVerifyEmail(user.Email); err != nil {
		return fmt.Errorf("failed to create verification email: %v", err)
	}

	return nil
}

func (s *service) Login(user models.User) (string, string, error) {
	userFromDB, err := s.repo.GetUserByEmail(user.Email, true)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password)); err != nil {
		return "", "", fmt.Errorf("invalid password")
	}

	accessToken, _, err := utils.GenerateJWT(userFromDB, "access")
	if err != nil {
		return "", "", err
	}

	refreshToken, jtiRefresh, err := utils.GenerateJWT(userFromDB, "refresh")
	if err != nil {
		return "", "", err
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
		return "", "", fmt.Errorf("failed to log token: %v", err)
	}

	return accessToken, refreshToken, nil
}

func (s *service) CreateVerifyEmail(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email, false)
	if err != nil || userFromDB.Email == "" {
		return fmt.Errorf("internal server error")
	}

	verifyEmail := models.VerifyEmail{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.repo.CreateVerifyEmail(verifyEmail); err != nil {
		return err
	}

	if err := utils.SendEmail(email, "Verify Email", "Click here to verify your email"); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *service) VerifyEmail(id string, c *gin.Context) error {
	user, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("failed to get user from context: %v", err)
	}

	_, err = uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	verifyEmail, err := s.repo.GetVerifyEmailByID(id)
	if err != nil || verifyEmail.ID == uuid.Nil {
		return fmt.Errorf("internal server error")
	}

	if verifyEmail.Email != user.Email {
		return fmt.Errorf("you are not authorized to verify this email")
	}

	if time.Now().After(verifyEmail.ExpiredAt) {
		return fmt.Errorf("token expired")
	}

	return s.repo.VerifyEmail(id)
}

func (s *service) ForgotPassword(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email, false)
	if err != nil || userFromDB.Email == "" {
		return fmt.Errorf("internal server error")
	}

	data := models.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateForgotPasswordEmail(data); err != nil {
		return fmt.Errorf("failed to create forgot password email: %v", err)
	}

	baseURL := os.Getenv("BASE_URL")
	if err := utils.SendEmail(email, "Password Reset", fmt.Sprintf(`
      Click here to reset your password: <a href="%s/reset-password?id=%s">Reset Password</a>`,
		baseURL, data.ID.String(),
	)); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *service) ResetPassword(id string, newPassword string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid token")
	}

	forgotPassword, err := s.repo.GetForgotPasswordByID(id)
	if err != nil || forgotPassword.ID == uuid.Nil {
		return fmt.Errorf("internal server error")
	}

	if time.Now().After(forgotPassword.ExpiredAt) {
		return fmt.Errorf("token expired")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password")
	}

	if err = s.repo.UpdateUserPassword(forgotPassword.UserID, string(hashedNewPassword)); err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	return s.repo.DeleteForgotPasswordByID(id)
}

func (s *service) InvalidateJWTTokens(oldJTI, newJTI string) error {
	if oldJTI == "" || newJTI == "" {
		return fmt.Errorf("old and new jti cannot be empty")
	}
	return s.repo.InvalidateTokenLog(oldJTI, newJTI)
}

func (s *service) IsTokenLogInvalidated(jti string) (bool, error) {
	if jti == "" {
		return false, fmt.Errorf("jti cannot be empty")
	}
	return s.repo.IsTokenLogInvalidated(jti)
}
