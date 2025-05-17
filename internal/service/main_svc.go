package service

import (
	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/Jonathan0823/auth-go/internal/repository"
)

type Service interface {
	// user service
	GetUserByID(id int) (dto.User, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}
