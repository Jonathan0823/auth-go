package domain

import "time"

type User struct {
	ID           int
	OAuthID      string
	Username     string
	AvatarURL    string
	Email        string
	PasswordHash string
	IsVerified   bool
	Provider     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UpdateUserCommand struct {
	ID        int
	Username  string
	AvatarURL string
	Email     string
}
