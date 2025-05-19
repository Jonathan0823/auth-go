package service

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
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
