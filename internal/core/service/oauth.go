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
	err := s.repo.Users().Create(ctx, user)
	if err != nil {
		// Conflict means the user already exists — that's fine, continue.
		if code := domain.ErrorCode(err); code != domain.ErrCodeConflict {
			return nil, domain.InternalServerError("failed to create user", err)
		}
	}

	userData, err := s.repo.Users().GetByEmail(ctx, user.Email, false)
	if err != nil || userData == nil {
		return nil, domain.InternalServerError("failed to retrieve user after creation", err)
	}
	return userData, nil
}
