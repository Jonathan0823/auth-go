package models

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	ImageURL   string `json:"image_url"`
	Email      string `json:"email"`
	Password   string `json:"password,omitempty"`
	IsVerified bool   `json:"is_verified"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
