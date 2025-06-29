package models

type User struct {
	ID         int    `json:"id"`
	OAuthID    string `json:"oauth_id,omitempty"`
	Username   string `json:"username" validate:"omitempty,min=3,max=30"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password,omitempty" validate:"required_without=OAuthID,min=8,max=100"`
	IsVerified bool   `json:"is_verified"`
	Provider   string `json:"provider,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type UpdateUserRequest struct {
	ID        int    `json:"id" validate:"required"`
	Username  string `json:"username" validate:"omitempty,min=3,max=30"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Email     string `json:"email" validate:"required,email"`
}
