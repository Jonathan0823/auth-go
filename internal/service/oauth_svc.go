package service

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
)

func (s *service) OAuthLogin(user models.User) (*models.User, error) {
	userData, err := s.repo.GetUserByEmail(user.Email, false)
	if err != nil {
		if err = s.repo.CreateUser(user); err != nil {
			return nil, errors.InternalServerError(fmt.Sprintf("failed to create user: %v", err))
		}
		userData, err = s.repo.GetUserByEmail(user.Email, false)
		if err != nil {
			return nil, errors.InternalServerError(fmt.Sprintf("failed to retrieve user after creation: %v", err))
		}
	}

	return userData, nil
}
