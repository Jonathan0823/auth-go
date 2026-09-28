package service

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/core/ratelimit"
)

func New(repo port.Repository, tokens port.TokenService, email port.EmailSender, hasher port.PasswordHasher, baseURL string, rateLimiter *ratelimit.RateLimiter) port.Service {
	users := repo.Users()
	return port.Service{
		Auth:  NewAuthService(repo, tokens, email, hasher, baseURL, rateLimiter),
		User:  NewUserService(users),
		OAuth: NewOAuthService(users),
	}
}
