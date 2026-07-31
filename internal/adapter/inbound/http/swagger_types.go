package http

import "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

type UserResponseEnvelope struct {
	Message string           `json:"message" example:"User retrieved successfully"`
	User    dto.UserResponse `json:"user"`
}

type UsersResponseEnvelope struct {
	Message string             `json:"message" example:"Users retrieved successfully"`
	Users   []dto.UserResponse `json:"users"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid input"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
