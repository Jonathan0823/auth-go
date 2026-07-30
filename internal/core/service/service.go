package service

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

func New(repo port.Repository, tokens port.TokenService, email port.EmailSender) port.Service {
	return port.Service{
		Auth:  NewAuthService(repo, tokens, email),
		User:  NewUserService(repo),
		OAuth: NewOAuthService(repo),
	}
}
