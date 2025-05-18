package service

import (
	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/gin-gonic/gin"
)

type Service interface {
	//auth service
	Register(user dto.User) error
	Login(user dto.User) (string, string, error)

	// user service
	GetUserByID(id int) (dto.User, error)
	GetUserByEmail(email string) (dto.User, error)
	GetAllUsers() ([]dto.User, error)
	UpdateUser(user dto.User, c *gin.Context) error
	DeleteUser(id int, c *gin.Context) error
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}
