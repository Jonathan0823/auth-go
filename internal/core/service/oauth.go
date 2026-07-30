package service

import (
	"context"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type oAuthService struct {
	repo port.Repository
}

func NewOAuthService(repo port.Repository) port.OAuthService {
	return &oAuthService{repo: repo}
}

func (s *oAuthService) OAuthLogin(ctx context.Context, user domain.User) (*domain.User, error) {
	// Try to create the user. If they already exist (unique violation), that's fine.
	createErr := s.repo.Users().Create(ctx, user)
	if createErr != nil {
		// If the repo returns an error, we try to look up the user by email anyway.
		// They likely already exist. Any other error will surface in GetByEmail.
	}

	userData, err := s.repo.Users().GetByEmail(ctx, user.Email, false)
	if err != nil || userData == nil {
		return nil, domain.InternalServerError("failed to retrieve user after creation", err)
	}

	return userData, nil
}
