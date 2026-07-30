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
		return fmt.Errorf("hash password: %w", err)
	}
	user.Password = hashed

	if err := s.repo.Users().Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if err := s.CreateVerifyEmail(ctx, user.Email); err != nil {
		return fmt.Errorf("create verification email: %w", err)
	}
	return nil
}

func (s *authService) Login(ctx context.Context, user domain.User) (string, string, error) {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, user.Email, true)
	if err != nil {
		return "", "", fmt.Errorf("get user by email: %w", err)
	}
	if userFromDB == nil {
		return "", "", fmt.Errorf("user not found: %w", domain.ErrNotFound)
	}

	if err := s.hasher.Compare(userFromDB.Password, user.Password); err != nil {
		return "", "", fmt.Errorf("invalid credentials: %w", domain.ErrUnauthenticated)
	}

	accessToken, _, err := s.tokens.GenerateAccessToken(*userFromDB)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, jtiRefresh, err := s.tokens.GenerateRefreshToken(*userFromDB)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
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
		return "", "", fmt.Errorf("create token log: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (s *authService) CreateVerifyEmail(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}
	if userFromDB == nil {
		return fmt.Errorf("user not found: %w", domain.ErrNotFound)
	}

	verifyEmail := domain.VerifyEmail{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.repo.Auth().CreateVerifyEmail(ctx, verifyEmail); err != nil {
		return fmt.Errorf("create verification email: %w", err)
	}
	if err := s.email.Send(email, "Verify Email", "Click here to verify your email"); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid verification token: %w", domain.ErrInvalidInput)
	}

	verifyEmail, err := s.repo.Auth().GetVerifyEmailByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get verification email: %w", err)
	}
	if verifyEmail.ID == uuid.Nil {
		return fmt.Errorf("verification token not found: %w", domain.ErrNotFound)
	}
	if time.Now().After(verifyEmail.ExpiredAt) {
		return fmt.Errorf("verification token expired: %w", domain.ErrInvalidInput)
	}
	if err := s.repo.Auth().VerifyEmail(ctx, id); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}
	if userFromDB == nil {
		return fmt.Errorf("user not found: %w", domain.ErrNotFound)
	}

	data := domain.ForgotPassword{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.repo.Auth().CreateForgotPasswordEmail(ctx, data); err != nil {
		return fmt.Errorf("create forgot password record: %w", err)
	}

	body := fmt.Sprintf(`
Click here to reset your password: <a href="%s/reset-password?id=%s">Reset Password</a>`, s.baseURL, data.ID.String())
	if err := s.email.Send(email, "Password Reset", body); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, tokenID, newPassword string) error {
	if _, err := uuid.Parse(tokenID); err != nil {
		return fmt.Errorf("invalid reset token: %w", domain.ErrInvalidInput)
	}

	hashed, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	return s.repo.WithTx(ctx, func(u port.UnitOfWork) error {
		forgotPassword, err := u.Auth().GetForgotPasswordByID(ctx, tokenID)
		if err != nil {
			return fmt.Errorf("get forgot password token: %w", err)
		}
		if forgotPassword.ID == uuid.Nil {
			return fmt.Errorf("forgot password token not found: %w", domain.ErrNotFound)
		}
		if time.Now().After(forgotPassword.ExpiredAt) {
			return fmt.Errorf("forgot password token expired: %w", domain.ErrInvalidInput)
		}
		if err := u.Auth().DeleteForgotPasswordByID(ctx, tokenID); err != nil {
			return fmt.Errorf("delete forgot password record: %w", err)
		}
		if err := u.Users().UpdatePassword(ctx, forgotPassword.UserID, hashed); err != nil {
			return fmt.Errorf("update user password: %w", err)
		}
		return nil
	})
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken, ip, userAgent string) (string, string, error) {
	claims, err := s.tokens.ValidateToken(refreshToken, "refresh")
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", domain.ErrUnauthenticated)
	}

	oldJTI, ok := claims["jti"].(string)
	if !ok {
		return "", "", fmt.Errorf("refresh token missing jti: %w", domain.ErrUnauthenticated)
	}
	isInvalidated, err := s.IsTokenLogInvalidated(ctx, oldJTI)
	if err != nil {
		return "", "", err
	}
	if isInvalidated {
		return "", "", fmt.Errorf("refresh token invalidated: %w", domain.ErrUnauthenticated)
	}

	username, okUsername := claims["username"].(string)
	email, okEmail := claims["email"].(string)
	if !okUsername || !okEmail {
		return "", "", fmt.Errorf("refresh token claims invalid: %w", domain.ErrUnauthenticated)
	}
	user := domain.User{Username: username, Email: email, IPAddress: ip, UserAgent: userAgent}

	newAccessToken, _, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}
	newRefreshToken, newJTI, err := s.tokens.GenerateRefreshToken(user)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	if err := s.InvalidateJWTTokens(ctx, oldJTI, newJTI); err != nil {
		return "", "", fmt.Errorf("invalidate old token: %w", err)
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *authService) InvalidateJWTTokens(ctx context.Context, oldJTI, newJTI string) error {
	if oldJTI == "" {
		return fmt.Errorf("old jti cannot be empty: %w", domain.ErrInvalidInput)
	}
	if err := s.repo.Auth().InvalidateTokenLog(ctx, oldJTI, newJTI); err != nil {
		return fmt.Errorf("invalidate token log: %w", err)
	}
	return nil
}

func (s *authService) IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, fmt.Errorf("jti cannot be empty: %w", domain.ErrInvalidInput)
	}
	invalidated, err := s.repo.Auth().IsTokenLogInvalidated(ctx, jti)
	if err != nil {
		return false, fmt.Errorf("check token invalidation: %w", err)
	}
	return invalidated, nil
}
