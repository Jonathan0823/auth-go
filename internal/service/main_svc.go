package service

import (
	"github.com/Jonathan0823/auth-go/internal/models"
	"github.com/Jonathan0823/auth-go/internal/repository"
	"github.com/gin-gonic/gin"
)

type Service interface {
	// auth service
	Register(user models.User) error
	Login(user models.User) (string, string, error)
	ForgotPassword(email string) error
	CreateVerifyEmail(email string) error
	VerifyEmail(id string, c *gin.Context) error
	ResetPassword(tokenStr string, newPassword string) error
	InvalidateJWTTokens(oldJTI, newJTI string) error
	IsTokenLogInvalidated(jti string) (bool, error)

	// oauth service
	OAuthLogin(user models.User) (*models.User, error)

	// user service
	GetUserByID(id int) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	UpdateUser(user models.UpdateUserRequest, c *gin.Context) error
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
