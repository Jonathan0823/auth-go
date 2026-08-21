package service

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

func New(repo port.Repository, tokens port.TokenService, email port.EmailSender, hasher port.PasswordHasher, baseURL string) port.Service {
	users := repo.Users()
	return port.Service{
		Auth:  NewAuthService(repo, tokens, email, hasher, baseURL),
		User:  NewUserService(users),
		OAuth: NewOAuthService(users),
	}
}
