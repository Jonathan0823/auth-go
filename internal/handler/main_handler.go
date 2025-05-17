package handler

import "github.com/Jonathan0823/auth-go/internal/service"

type MainHandler struct {
	svc service.Service
}

func NewMainHandler(svc service.Service) *MainHandler {
	return &MainHandler{
		svc: svc,
	}
}
