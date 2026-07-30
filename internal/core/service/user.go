package service

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("user not found: %w", domain.ErrNotFound)
	}
	return data, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	data, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("user not found: %w", domain.ErrNotFound)
	}
	return data, nil
}

func (s *userService) GetAll(ctx context.Context) ([]*domain.User, error) {
	data, err := s.repo.Users().GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	return data, nil
}

func (s *userService) Update(ctx context.Context, currentUserID int, user domain.UpdateUserRequest) error {
	if currentUserID != user.ID {
		return fmt.Errorf("update user %d forbidden: %w", user.ID, domain.ErrForbidden)
	}
	if err := s.repo.Users().Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, id int, requestingUserID int) error {
	if requestingUserID != id {
		return fmt.Errorf("delete user %d forbidden: %w", id, domain.ErrForbidden)
	}
	if err := s.repo.Users().Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
