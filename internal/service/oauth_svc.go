package service

import (
	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
)

func (s *service) OAuthLogin(user models.User) (*models.User, error) {
	userData, err := s.repo.GetUserByEmail(user.Email, false)
	if err != nil {
		if err = s.repo.CreateUser(user); err != nil {
			return nil, errors.InternalServerError("failed to create user", err)
		}
		userData, err = s.repo.GetUserByEmail(user.Email, false)
		if err != nil {
			return nil, errors.InternalServerError("failed to retrieve user after creation", err)
		}
	}

	return userData, nil
}
