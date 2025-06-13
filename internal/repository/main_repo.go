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

	// user repo
	GetUserByID(id int) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	CreateUser(user models.User) error
	GetAllUsers() ([]models.User, error)
	UpdateUser(user models.User) error
	DeleteUser(id int) error
	UpdateUserPassword(id int, newPassword string) error
	GetPasswordByEmail(email string) (string, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}
