// Package repository provides methods for interacting with the database related to email verification and password reset functionalities
package repository

import (
	"database/sql"

	"github.com/Jonathan0823/auth-go/internal/models"
)

type Repository interface {
	// auth repo
	CreateVerifyEmail(verifyEmail models.VerifyEmail) error
	GetVerifyEmailByID(id string) (models.VerifyEmail, error)
	VerifyEmail(id string) error
	CreateForgotPasswordEmail(data models.ForgotPassword) error
	GetForgotPasswordByID(id string) (models.ForgotPassword, error)
	DeleteForgotPasswordByID(id string) error
	CreateTokenLog(tokenLog models.TokenLog) error
	GetTokenLogByJTI(jti string) (models.TokenLog, error)
	InvalidateTokenLog(oldJti, newJti string) error
	IsTokenLogInvalidated(jti string) (bool, error)

	// user repo
	GetUserByID(id int) (models.User, error)
	GetUserByEmail(email string, includePassword bool) (models.User, error)
	CreateUser(user models.User) error
	GetAllUsers() ([]models.User, error)
	UpdateUser(user models.UpdateUserRequest) error
	DeleteUser(id int) error
	UpdateUserPassword(id int, newPassword string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}
