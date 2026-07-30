package http

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type Handler struct {
	Svc    port.Service
	Tokens port.TokenService
}

func NewHandler(svc port.Service, tokens port.TokenService) *Handler {
	return &Handler{Svc: svc, Tokens: tokens}
}
