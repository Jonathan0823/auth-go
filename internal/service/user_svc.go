package service

import (
	"context"

	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, user models.UpdateUserRequest, c *gin.Context) error
	DeleteUser(ctx context.Context, id int, c *gin.Context) error
}

type userService struct {
	repo repository.Repository
}

func NewUserService(repo repository.Repository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	data, err := s.repo.Users().GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError("failed to get user by id", err)
	}
	if data == nil {
		return nil, errors.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	data, err := s.repo.Users().GetUserByEmail(ctx, email, false)
	if err != nil {
		return nil, errors.InternalServerError("failed to get user by email", err)
	}
	if data == nil {
		return nil, errors.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	data, err := s.repo.Users().GetAllUsers(ctx)
	if err != nil {
		return nil, errors.InternalServerError("failed to get all users", err)
	}
	return data, nil
}

func (s *userService) UpdateUser(ctx context.Context, user models.UpdateUserRequest, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return errors.Unauthorized("user is not found", err)
	}

	if currentUser.ID != user.ID {
		return errors.Forbidden("you are not authorized to update this user", nil)
	}

	if err := s.repo.Users().UpdateUser(ctx, user); err != nil {
		return errors.InternalServerError("failed to update user", err)
	}
	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id int, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return errors.Unauthorized("user is not found", err)
	}

	if currentUser.ID != id {
		return errors.Forbidden("you are not authorized to delete this user", nil)
	}

	if err := s.repo.Users().DeleteUser(ctx, id); err != nil {
		return errors.InternalServerError("failed to delete user", err)
	}
	return nil
}
