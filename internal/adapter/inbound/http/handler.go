package http

import (
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type Handler struct {
	Svc port.Service
}

func NewHandler(svc port.Service) *Handler {
	return &Handler{Svc: svc}
}
