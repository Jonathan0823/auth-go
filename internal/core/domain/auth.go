package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerifyEmail struct {
	ID        uuid.UUID
	UserID    int
	Email     string
	ExpiredAt time.Time
	CreatedAt time.Time
}

type ForgotPassword struct {
	ID        uuid.UUID
	UserID    int
	Email     string
	ExpiredAt time.Time
	CreatedAt time.Time
}

type TokenLog struct {
	ID               uuid.UUID
	UserID           int
	JTI              string
	RefreshedFromJTI *string
	InvalidatedAt    *time.Time
	ExpiredAt        time.Time
	CreatedAt        time.Time
	IPAddress        string
	UserAgent        string
}
