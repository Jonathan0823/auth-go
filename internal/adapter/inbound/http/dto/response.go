package dto

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

type UserResponseEnvelope struct {
	Message string       `json:"message" example:"User retrieved successfully"`
	User    UserResponse `json:"user"`
}

type UsersResponseEnvelope struct {
	Message string         `json:"message" example:"Users retrieved successfully"`
	Users   []UserResponse `json:"users"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid input"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
