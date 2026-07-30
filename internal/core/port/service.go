package port

import (
	"context"
	"net/http"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

type AuthService interface {
	Register(ctx context.Context, user domain.User) error
	Login(ctx context.Context, user domain.User) (accessToken, refreshToken string, err error)
	ForgotPassword(ctx context.Context, email string) error
	CreateVerifyEmail(ctx context.Context, email string) error
	VerifyEmail(ctx context.Context, id string) error
	ResetPassword(ctx context.Context, tokenID, newPassword string) error
	RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (accessToken, refreshTokenNew string, err error)
	InvalidateJWTTokens(ctx context.Context, oldJTI, newJTI string) error
	IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error)
}

type UserService interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetAll(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, currentUserID int, user domain.UpdateUserCommand) error
	Delete(ctx context.Context, id int, requestingUserID int) error
}

type OAuthService interface {
	BeginAuth(w http.ResponseWriter, r *http.Request, provider string)
	OAuthLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, provider string) (*domain.User, error)
}

type Service struct {
	Auth  AuthService
	User  UserService
	OAuth OAuthService
}
