// Package models provides data models for the application
package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

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

type TokenLog struct {
	ID               uuid.UUID  `json:"id"`
	UserID           int        `json:"user_id"`
	JTI              string     `json:"jti"`
	RefreshedFromJTI *string    `json:"refreshed_from_jti"`
	InvalidatedAt    *time.Time `json:"invalidated_at"`
	ExpiredAt        time.Time  `json:"expired_at"`
	CreatedAt        time.Time  `json:"created_at"`
	IPAddress        string     `json:"ip_address"`
	UserAgent        string     `json:"user_agent"`
}
