package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string, includePassword bool) (*domain.User, error)
	Create(ctx context.Context, user domain.User) error
	GetAll(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, user domain.UpdateUserCommand) error
	Delete(ctx context.Context, id int) error
	UpdatePassword(ctx context.Context, id int, newPassword string) error
}

type AuthRepository interface {
	CreateVerifyEmail(ctx context.Context, verifyEmail domain.VerifyEmail) error
	GetVerifyEmailByID(ctx context.Context, id string) (domain.VerifyEmail, error)
	VerifyEmail(ctx context.Context, id string) error
	CreateForgotPasswordEmail(ctx context.Context, data domain.ForgotPassword) error
	GetForgotPasswordByID(ctx context.Context, id string) (domain.ForgotPassword, error)
	DeleteForgotPasswordByID(ctx context.Context, id string) error
	CreateRefreshToken(ctx context.Context, rt domain.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash []byte) (*domain.RefreshToken, error)
	UseRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error
}

type UnitOfWork interface {
	Users() UserRepository
	Auth() AuthRepository
	Commit() error
	Rollback() error
}

type Repository interface {
	Begin(ctx context.Context) (UnitOfWork, error)
	WithTx(ctx context.Context, fn func(u UnitOfWork) error) error
	Users() UserRepository
	Auth() AuthRepository
}
