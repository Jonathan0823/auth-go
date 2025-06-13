package models

import (
	"time"

	"github.com/google/uuid"
)

type VerifyEmail struct {
	ID        uuid.UUID `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email" validate:"required,email"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ForgotPassword struct {
	ID        uuid.UUID `json:"id"`
	UserID    int       `json:"user_id"`
	Email     string    `json:"email" validate:"required,email"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ResetPasswordRequest struct {
	ID       string `json:"id" validate:"required,uuid"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}
