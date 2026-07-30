package port

import (
	"context"

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
	CreateTokenLog(ctx context.Context, tokenLog domain.TokenLog) error
	GetTokenLogByJTI(ctx context.Context, jti string) (domain.TokenLog, error)
	InvalidateTokenLog(ctx context.Context, oldJTI, newJTI string) error
	IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error)
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
