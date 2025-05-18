package repository

import (
	"database/sql"

	"github.com/Jonathan0823/auth-go/internal/dto"
)

type Repository interface {
	//user repo
	GetUserByID(id int) (dto.User, error)
	GetUserByEmail(email string) (dto.User, error)
	CreateUser(user dto.User) error
	GetAllUsers() ([]dto.User, error)
	UpdateUser(user dto.User) error
	DeleteUser(id int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{
		db: db,
	}
}
