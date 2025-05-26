package models

type User struct {
	ID         int    `json:"id"`
	OAuthID    string `json:"oauth_id,omitempty"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Email      string `json:"email"`
	Password   string `json:"password,omitempty"`
	IsVerified bool   `json:"is_verified"`
	Provider   string `json:"provider,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
