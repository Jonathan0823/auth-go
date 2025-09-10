package service

import (
	"github.com/Jonathan0823/auth-go/internal/repository"
)

type Service interface {
	User() UserService
	OAuth() OAuthService
	Auth() AuthService
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) User() UserService {
	return NewUserService(s.repo)
}

func (s *service) OAuth() OAuthService {
	return NewOAuthService(s.repo)
}

func (s *service) Auth() AuthService {
	return NewAuthService(s.repo)
}
