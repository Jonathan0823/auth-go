package dto

import (
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

type UpdateUserRequest struct {
	ID        int    `json:"id" validate:"required"`
	Username  string `json:"username" validate:"omitempty,min=3,max=30"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Email     string `json:"email" validate:"required,email"`
}

type UserResponse struct {
	ID         int       `json:"id"`
	OAuthID    string    `json:"oauth_id,omitempty"`
	Username   string    `json:"username"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	Provider   string    `json:"provider,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func UserResponseFromDomain(user *domain.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		OAuthID:    user.OAuthID,
		Username:   user.Username,
		AvatarURL:  user.AvatarURL,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		Provider:   user.Provider,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}

func UserResponsesFromDomain(users []*domain.User) []UserResponse {
	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, UserResponseFromDomain(user))
	}
	return responses
}
