package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type oAuthService struct {
	users port.UserRepository
}

func NewOAuthService(users port.UserRepository) port.OAuthService {
	return &oAuthService{users: users}
}

func (s *oAuthService) Login(ctx context.Context, profile domain.OAuthProfile) (*domain.User, error) {
	if profile.UserID == "" {
		return nil, fmt.Errorf("oauth profile user ID is required: %w", domain.ErrInvalidInput)
	}
	if profile.Email == "" {
		return nil, fmt.Errorf("oauth profile email is required: %w", domain.ErrInvalidInput)
	}
	if profile.Provider == "" {
		return nil, fmt.Errorf("oauth profile provider is required: %w", domain.ErrInvalidInput)
	}

	user := domain.User{
		OAuthID:   profile.UserID,
		Email:     profile.Email,
		Username:  profile.Name,
		Provider:  profile.Provider,
		AvatarURL: profile.AvatarURL,
	}
	if err := s.users.Create(ctx, user); err != nil && !errors.Is(err, domain.ErrConflict) {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}

	userData, err := s.users.GetByEmail(ctx, user.Email, false)
	if err != nil {
		return nil, fmt.Errorf("get oauth user: %w", err)
	}
	if userData == nil {
		return nil, fmt.Errorf("oauth user not found: %w", domain.ErrNotFound)
	}
	return userData, nil
}
