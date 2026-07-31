package http

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

type Handler struct {
	Svc         port.Service
	Tokens      port.TokenService // used by middleware
	Audit       *platform.AuditLogger
	RateLimiter *platform.RateLimiter
}

func NewHandler(svc port.Service, tokens port.TokenService) *Handler {
	return &Handler{Svc: svc, Tokens: tokens}
}
