package service

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (s *service) GetUserByID(id int) (models.User, error) {
	data, err := s.repo.GetUserByID(id)
	if err != nil {
		return models.User{}, fmt.Errorf("user not found: %w", err)
	}
	return data, nil
}

func (s *service) GetUserByEmail(email string) (models.User, error) {
	data, err := s.repo.GetUserByEmail(email, false)
	if err != nil {
		return models.User{}, fmt.Errorf("user not found: %w", err)
	}
	return data, nil
}

func (s *service) GetAllUsers() ([]models.User, error) {
	data, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return data, nil
}

func (s *service) UpdateUser(user models.UpdateUserRequest, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("user is not found")
	}

	if currentUser.ID != user.ID {
		return fmt.Errorf("you are not authorized to update this user")
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (s *service) DeleteUser(id int, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("user is not found")
	}

	if currentUser.ID != id {
		return fmt.Errorf("you are not authorized to update this user")
	}

	if err := s.repo.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
