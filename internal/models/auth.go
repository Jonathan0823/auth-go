package models

import (
	"time"

	"github.com/google/uuid"
)

type VerifyEmail struct {
	ID        uuid.UUID `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email" binding:"required,email"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ForgotPassword struct {
	ID        uuid.UUID `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email" binding:"required,email"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}
