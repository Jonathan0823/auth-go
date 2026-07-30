package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type authService struct {
	repo    port.Repository
	tokens  port.TokenService
	email   port.EmailSender
	hasher  port.PasswordHasher
	baseURL string
}

func NewAuthService(repo port.Repository, tokens port.TokenService, email port.EmailSender, hasher port.PasswordHasher, baseURL string) port.AuthService {
	return &authService{
		repo:    repo,
		tokens:  tokens,
		email:   email,
		hasher:  hasher,
		baseURL: baseURL,
	}
}

func (s *authService) Register(ctx context.Context, user domain.User) error {
	hashed, err := s.hasher.Hash(user.Password)
	if err != nil {
		return domain.InternalServerError("failed to hash password", err)
	}
	user.Password = hashed

	if err := s.repo.Users().Create(ctx, user); err != nil {
		return domain.InternalServerError("failed to create user", err)
	}

	if err := s.CreateVerifyEmail(ctx, user.Email); err != nil {
		return domain.InternalServerError("failed to create verification email", err)
	}
	return nil
}

func (s *authService) Login(ctx context.Context, user domain.User) (string, string, error) {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, user.Email, true)
	if err != nil {
		return "", "", domain.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return "", "", domain.NotFound("user not found", nil)
	}

	if err := s.hasher.Compare(userFromDB.Password, user.Password); err != nil {
		return "", "", domain.Unauthorized("invalid credentials", err)
	}

	accessToken, _, err := s.tokens.GenerateAccessToken(*userFromDB)
	if err != nil {
		return "", "", domain.InternalServerError("failed to generate access token", err)
	}

	refreshToken, jtiRefresh, err := s.tokens.GenerateRefreshToken(*userFromDB)
	if err != nil {
		return "", "", domain.InternalServerError("failed to generate refresh token", err)
	}

	tokenLog := domain.TokenLog{
		ID:               uuid.New(),
		UserID:           userFromDB.ID,
		JTI:              jtiRefresh,
		RefreshedFromJTI: nil,
		InvalidatedAt:    nil,
		ExpiredAt:        time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:        time.Now(),
		IPAddress:        user.IPAddress,
		UserAgent:        user.UserAgent,
	}

	if err := s.repo.Auth().CreateTokenLog(ctx, tokenLog); err != nil {
		return "", "", domain.InternalServerError("failed to create token log", err)
	}

	return accessToken, refreshToken, nil
}

func (s *authService) CreateVerifyEmail(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return domain.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return domain.NotFound("user not found", nil)
	}

	verifyEmail := domain.VerifyEmail{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.repo.Auth().CreateVerifyEmail(ctx, verifyEmail); err != nil {
		return domain.InternalServerError("failed to create verification email", err)
	}

	if err := s.email.Send(email, "Verify Email", "Click here to verify your email"); err != nil {
		return domain.InternalServerError("failed to send verification email", err)
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.BadRequest("invalid token", err)
	}

	verifyEmail, err := s.repo.Auth().GetVerifyEmailByID(ctx, id)
	if err != nil {
		return domain.InternalServerError("internal server error", err)
	}
	if verifyEmail.ID == uuid.Nil {
		return domain.NotFound("verification token not found", nil)
	}

	if time.Now().After(verifyEmail.ExpiredAt) {
		return domain.BadRequest("token expired", nil)
	}

	if err := s.repo.Auth().VerifyEmail(ctx, id); err != nil {
		return domain.InternalServerError("failed to verify email", err)
	}
	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return domain.InternalServerError("failed to get user by email", err)
	}
	if userFromDB == nil {
		return domain.NotFound("user not found", nil)
	}

	data := domain.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}

	if err := s.repo.Auth().CreateForgotPasswordEmail(ctx, data); err != nil {
		return domain.InternalServerError("failed to create forgot password record", err)
	}

	body := fmt.Sprintf(`
Click here to reset your password: <a href="%s/reset-password?id=%s">Reset Password</a>`,
		s.baseURL, data.ID.String())
	if err := s.email.Send(email, "Password Reset", body); err != nil {
		return domain.InternalServerError("failed to send password reset email", err)
	}
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	if _, err := uuid.Parse(tokenID); err != nil {
		return domain.BadRequest("invalid token", err)
	}

	hashed, err := s.hasher.Hash(newPassword)
	if err != nil {
		return domain.InternalServerError("failed to hash new password", err)
	}

	return s.repo.WithTx(ctx, func(u port.UnitOfWork) error {
		forgotPassword, err := u.Auth().GetForgotPasswordByID(ctx, tokenID)
		if err != nil {
			return domain.InternalServerError("internal server error", err)
		}
		if forgotPassword.ID == uuid.Nil {
			return domain.NotFound("forgot password token not found", nil)
		}

		if time.Now().After(forgotPassword.ExpiredAt) {
			return domain.BadRequest("token expired", nil)
		}

		if err := u.Auth().DeleteForgotPasswordByID(ctx, tokenID); err != nil {
			return domain.InternalServerError("failed to delete forgot password record", err)
		}

		if err = u.Users().UpdatePassword(ctx, forgotPassword.UserID, hashed); err != nil {
			return domain.InternalServerError("failed to update user password", err)
		}
		return nil
	})
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (string, string, error) {
	claims, err := s.tokens.ValidateToken(refreshToken, "refresh")
	if err != nil {
		return "", "", domain.Unauthorized("invalid refresh token", err)
	}

	oldJTI := claims["jti"].(string)
	isInvalidated, err := s.IsTokenLogInvalidated(ctx, oldJTI)
	if err != nil || isInvalidated {
		return "", "", domain.Unauthorized("invalidated refresh token", err)
	}

	user := domain.User{
		Username:  claims["username"].(string),
		Email:     claims["email"].(string),
		IPAddress: ip,
		UserAgent: userAgent,
	}

	newAccessToken, _, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return "", "", domain.InternalServerError("failed to generate access token", err)
	}

	newRefreshToken, newJTI, err := s.tokens.GenerateRefreshToken(user)
	if err != nil {
		return "", "", domain.InternalServerError("failed to generate refresh token", err)
	}

	if err := s.InvalidateJWTTokens(ctx, oldJTI, newJTI); err != nil {
		return "", "", domain.InternalServerError("failed to invalidate old tokens", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) InvalidateJWTTokens(ctx context.Context, oldJTI, newJTI string) error {
	if oldJTI == "" {
		return domain.BadRequest("oldJTI cannot be empty", nil)
	}
	if err := s.repo.Auth().InvalidateTokenLog(ctx, oldJTI, newJTI); err != nil {
		return domain.InternalServerError("failed to invalidate token log", err)
	}
	return nil
}

func (s *authService) IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, domain.BadRequest("jti cannot be empty", nil)
	}
	invalidated, err := s.repo.Auth().IsTokenLogInvalidated(ctx, jti)
	if err != nil {
		return false, domain.InternalServerError("failed to check if token log is invalidated", err)
	}
	return invalidated, nil
}
