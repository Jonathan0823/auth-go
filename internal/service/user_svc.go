package service

import (
	"fmt"

	"github.com/Jonathan0823/auth-go/internal/dto"
	"github.com/Jonathan0823/auth-go/utils"
	"github.com/gin-gonic/gin"
)

func (s *service) GetUserByID(id int) (dto.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *service) GetUserByEmail(email string) (dto.User, error) {
	return s.repo.GetUserByEmail(email)
}

func (s *service) GetAllUsers() ([]dto.User, error) {
	return s.repo.GetAllUsers()
}

func (s *service) UpdateUser(user dto.User, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("User is not found")
	}

	if currentUser.ID != user.ID {
		return fmt.Errorf("You are not authorized to update this user")
	}

	return s.repo.UpdateUser(user)
}

func (s *service) DeleteUser(id int, c *gin.Context) error {
	currentUser, err := utils.GetUser(c)
	if err != nil {
		return fmt.Errorf("User is not found")
	}

	if currentUser.ID != id {
		return fmt.Errorf("You are not authorized to update this user")
	}

	return s.repo.DeleteUser(id)
}
