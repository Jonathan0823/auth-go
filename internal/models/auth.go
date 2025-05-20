package models

import (
	"time"

	"github.com/google/uuid"
)

type VerifyEmail struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email" binding:"required,email"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}
