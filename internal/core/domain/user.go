package domain

import "time"

type User struct {
	ID         int
	OAuthID    string
	Username   string
	AvatarURL  string
	Email      string
	Password   string
	IsVerified bool
	Provider   string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type UpdateUserCommand struct {
	ID        int
	Username  string
	AvatarURL string
	Email     string
}
