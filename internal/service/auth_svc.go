package service

import (
	"fmt"
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

	userFromDB, err := s.repo.GetUserByEmail(user.Email)
	if userFromDB.Email != "" {
		return fmt.Errorf("User with this email already exists")
	}

	user.Password = string(hashedPassword)
	if err := s.repo.CreateUser(user); err != nil {
		return err
	}

	if err := s.CreateVerifyEmail(user.Email); err != nil {
		return fmt.Errorf("Failed to create verification email: %v", err)
	}

	return nil
}

func (s *service) Login(user models.User) (string, string, error) {
	userFromDB, err := s.repo.GetUserByEmail(user.Email)
	if err != nil {
		return "", "", fmt.Errorf("User not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userFromDB.Password), []byte(user.Password)); err != nil {
		return "", "", fmt.Errorf("Invalid password")
	}

	access_token, err := utils.GenerateJWT(userFromDB, "access")
	if err != nil {
		return "", "", err
	}

	refresh_token, err := utils.GenerateJWT(userFromDB, "refresh")
	if err != nil {
		return "", "", err
	}

	return access_token, refresh_token, nil
}

func (s *service) CreateVerifyEmail(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email)
	if err != nil || userFromDB.Email == "" {
		return fmt.Errorf("Internal server error")
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
		return fmt.Errorf("Failed to send email: %v", err)
	}

	return nil
}

func (s *service) VerifyEmail(id string, c *gin.Context) error {
	user, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("Failed to get user from context: %v", err)
	}

	_, err = uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("Invalid token")
	}

	verifyEmail, err := s.repo.GetVerifyEmailByID(id)
	if err != nil || verifyEmail.ID == uuid.Nil {
		return fmt.Errorf("Internal server error")
	}

	if verifyEmail.Email != user.Email {
		return fmt.Errorf("You are not authorized to verify this email")
	}

	if time.Now().After(verifyEmail.ExpiredAt) {
		return fmt.Errorf("Token expired")
	}

	return s.repo.VerifyEmail(id)
}

func (s *service) ForgotPassword(email string) error {
	userFromDB, err := s.repo.GetUserByEmail(email)
	if err != nil || userFromDB.Email == "" {
		return fmt.Errorf("Internal server error")
	}

	data := models.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.CreateForgotPasswordEmail(data); err != nil {
		return fmt.Errorf("Failed to create forgot password email: %v", err)
	}

	if err := utils.SendEmail(email, "Password Reset", "Click here to reset your password"); err != nil {
		return fmt.Errorf("Failed to send email: %v", err)
	}

	return nil
}

func (s *service) ResetPassword(tokenStr string, newPassword string) error {
	if _, err := uuid.Parse(tokenStr); err != nil {
		return fmt.Errorf("Invalid token")
	}

	forgotPassword, err := s.repo.GetForgotPasswordByID(tokenStr)
	if err != nil || forgotPassword.ID == uuid.Nil {
		return fmt.Errorf("Internal server error")
	}

	if time.Now().After(forgotPassword.ExpiredAt) {
		return fmt.Errorf("Token expired")
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to hash new password")
	}

	if err = s.repo.UpdateUserPassword(forgotPassword.UserID, string(hashedNewPassword)); err != nil {
		return fmt.Errorf("Failed to update password: %v", err)
	}

	return s.repo.DeleteForgotPasswordByID(tokenStr)
}
