package service

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
)

func (s *service) OAuthLogin(user models.User) (*models.User, error) {
	userData, err := s.repo.GetUserByEmail(user.Email, false)
	if err != nil {
		if err = s.repo.CreateUser(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		userData, err = s.repo.GetUserByEmail(user.Email, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get user by email after creation: %w", err)
		}
	}

	return userData, nil
}
