package service

import (
	"github.com/Jonathan0823/auth-go/internal/repository"
)

type service struct {
	oauth OAuthService
}

func NewService(repo repository.Repository) *service {
	oauthSvc := NewOAuthService(repo)
	return &service{
		oauth: oauthSvc,
	}
}
