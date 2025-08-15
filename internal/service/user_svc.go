package service

import (
	"github.com/Jonathan0823/auth-go/internal/errors"
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (s *service) GetUserByID(id int) (*models.User, error) {
	data, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, errors.InternalServerError("failed to get user by id", err)
	}
	if data == nil {
		return nil, errors.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *service) GetUserByEmail(email string) (*models.User, error) {
	data, err := s.repo.GetUserByEmail(email, false)
	if err != nil {
		return nil, errors.InternalServerError("failed to get user by email", err)
	}
	if data == nil {
		return nil, errors.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *service) GetAllUsers() ([]*models.User, error) {
	data, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, errors.InternalServerError("failed to get all users", err)
	}
	return data, nil
}

func (s *service) UpdateUser(user models.UpdateUserRequest, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return errors.Unauthorized("user is not found", err)
	}

	if currentUser.ID != user.ID {
		return errors.Forbidden("you are not authorized to update this user", nil)
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return errors.InternalServerError("failed to update user", err)
	}
	return nil
}

func (s *service) DeleteUser(id int, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return errors.Unauthorized("user is not found", err)
	}

	if currentUser.ID != id {
		return errors.Forbidden("you are not authorized to delete this user", nil)
	}

	if err := s.repo.DeleteUser(id); err != nil {
		return errors.InternalServerError("failed to delete user", err)
	}
	return nil
}
