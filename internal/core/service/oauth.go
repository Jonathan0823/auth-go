package service

import (
	"context"
	"errors"
	"fmt"

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
	if err := s.repo.Users().Create(ctx, user); err != nil && !errors.Is(err, domain.ErrConflict) {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}

	userData, err := s.repo.Users().GetByEmail(ctx, user.Email, false)
	if err != nil {
		return nil, fmt.Errorf("get oauth user: %w", err)
	}
	if userData == nil {
		return nil, fmt.Errorf("oauth user not found: %w", domain.ErrNotFound)
	}
	return userData, nil
}
