package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type oAuthService struct {
	repo  port.Repository
	oauth port.OAuthClient
}

func NewOAuthService(repo port.Repository, oauth port.OAuthClient) port.OAuthService {
	return &oAuthService{repo: repo, oauth: oauth}
}

func (s *oAuthService) BeginAuth(w http.ResponseWriter, r *http.Request, provider string) {
	s.oauth.BeginAuth(w, r, provider)
}

func (s *oAuthService) OAuthLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, provider string) (*domain.User, error) {
	profile, err := s.oauth.CompleteAuth(w, r, provider)
	if err != nil {
		return nil, fmt.Errorf("complete oauth authentication: %w", domain.ErrUnauthenticated)
	}

	user := domain.User{
		OAuthID:   profile.UserID,
		Email:     profile.Email,
		Username:  profile.Name,
		Provider:  profile.Provider,
		AvatarURL: profile.AvatarURL,
	}
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
