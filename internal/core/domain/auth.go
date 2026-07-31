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

type RefreshToken struct {
	ID        uuid.UUID
	UserID    int
	TokenHash []byte
	FamilyID  uuid.UUID
	ParentID  *uuid.UUID
	ExpiredAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	IPAddress string
	UserAgent string
}
