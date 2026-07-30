package service

import (
	"context"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type userService struct {
	repo port.Repository
}

func NewUserService(repo port.Repository) port.UserService {
	return &userService{repo: repo}
}

func (s *userService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	data, err := s.repo.Users().GetByID(ctx, id)
	if err != nil {
		return nil, domain.InternalServerError("failed to get user by id", err)
	}
	if data == nil {
		return nil, domain.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	data, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return nil, domain.InternalServerError("failed to get user by email", err)
	}
	if data == nil {
		return nil, domain.NotFound("user not found", nil)
	}
	return data, nil
}

func (s *userService) GetAll(ctx context.Context) ([]*domain.User, error) {
	data, err := s.repo.Users().GetAll(ctx)
	if err != nil {
		return nil, domain.InternalServerError("failed to get all users", err)
	}
	return data, nil
}

func (s *userService) Update(ctx context.Context, currentUserID int, user domain.UpdateUserRequest) error {
	if currentUserID != user.ID {
		return domain.Forbidden("you are not authorized to update this user", nil)
	}
	if err := s.repo.Users().Update(ctx, user); err != nil {
		return domain.InternalServerError("failed to update user", err)
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, id int, requestingUserID int) error {
	if requestingUserID != id {
		return domain.Forbidden("you are not authorized to delete this user", nil)
	}
	if err := s.repo.Users().Delete(ctx, id); err != nil {
		return domain.InternalServerError("failed to delete user", err)
	}
	return nil
}
