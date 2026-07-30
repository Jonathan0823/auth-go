package http

import "github.com/Jonathan0823/auth-go/internal/core/port"

type Handler struct {
	Svc    port.Service
	Tokens port.TokenService
	OAuth  port.OAuthClient
}

func NewHandler(svc port.Service, tokens port.TokenService, oauth port.OAuthClient) *Handler {
	return &Handler{Svc: svc, Tokens: tokens, OAuth: oauth}
}
