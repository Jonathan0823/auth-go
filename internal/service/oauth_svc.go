package service

import (
	"context"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/Jonathan0823/auth-go/utils"
)

type OAuthService interface {
	OAuthLogin(ctx context.Context, user models.User) (*models.User, error)
}

type oAuthService struct {
	repo repository.Repository
}

func NewOAuthService(repo repository.Repository) OAuthService {
	return &oAuthService{
		repo: repo,
	}
}

func (s *oAuthService) OAuthLogin(ctx context.Context, user models.User) (*models.User, error) {
	if err := s.repo.Users().CreateUser(ctx, user); err != nil {
		if !utils.IsPGUniqueViolation(err) {
			return nil, errors.InternalServerError("failed to create user", err)
		}
	}
	userData, err := s.repo.Users().GetUserByEmail(ctx, user.Email, false)
	if err != nil || userData == nil {
		return nil, errors.InternalServerError("failed to retrieve user after creation", err)
	}

	return userData, nil
}
