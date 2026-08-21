package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

var ErrRefreshTokenReused = errors.New("refresh token reused")

const (
	errGetUserByEmail = "get user by email: %w"
	errUserNotFound   = "user not found: %w"
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

	var verification domain.VerifyEmail
	if err := s.repo.WithTx(ctx, func(u port.UnitOfWork) error {
		if err := u.Users().Create(ctx, user); err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		verification, err = createVerification(ctx, u.Users(), u.Auth(), user.Email)
		return err
	}); err != nil {
		return fmt.Errorf("register user: %w", err)
	}
	return s.sendVerificationEmail(verification)
}

func (s *authService) Login(ctx context.Context, user domain.User) (string, string, error) {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, user.Email, true)
	if err != nil {
		return "", "", fmt.Errorf(errGetUserByEmail, err)
	}
	if userFromDB == nil {
		return "", "", fmt.Errorf(errUserNotFound, domain.ErrNotFound)
	}

	if err := s.hasher.Compare(userFromDB.Password, user.Password); err != nil {
		return "", "", fmt.Errorf("invalid credentials: %w", domain.ErrUnauthenticated)
	}

	accessToken, _, err := s.tokens.GenerateAccessToken(*userFromDB)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	rawRefresh, hmacHash, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	rt := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userFromDB.ID,
		TokenHash: hmacHash,
		FamilyID:  uuid.New(),
		ParentID:  nil,
		ExpiredAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		IPAddress: user.IPAddress,
		UserAgent: user.UserAgent,
	}
	if err := s.repo.Auth().CreateRefreshToken(ctx, rt); err != nil {
		return "", "", fmt.Errorf("create refresh token: %w", err)
	}

	return accessToken, rawRefresh, nil
}

func (s *authService) CreateVerifyEmail(ctx context.Context, email string) error {
	verification, err := createVerification(ctx, s.repo.Users(), s.repo.Auth(), email)
	if err != nil {
		return err
	}
	return s.sendVerificationEmail(verification)
}

func createVerification(ctx context.Context, users port.UserRepository, auth port.AuthRepository, email string) (domain.VerifyEmail, error) {
	user, err := users.GetByEmail(ctx, email, false)
	if err != nil {
		return domain.VerifyEmail{}, fmt.Errorf(errGetUserByEmail, err)
	}
	if user == nil {
		return domain.VerifyEmail{}, fmt.Errorf(errUserNotFound, domain.ErrNotFound)
	}

	verification := domain.VerifyEmail{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     email,
		ExpiredAt: time.Now().Add(time.Hour),
	}
	if err := auth.CreateVerifyEmail(ctx, verification); err != nil {
		return domain.VerifyEmail{}, fmt.Errorf("create verification email: %w", err)
	}
	return verification, nil
}

func (s *authService) sendVerificationEmail(verification domain.VerifyEmail) error {
	body := fmt.Sprintf(`<a href="%s/api/auth/verify/email?id=%s">Verify email</a>`, s.baseURL, verification.ID)
	if err := s.email.Send(verification.Email, "Verify Email", body); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid verification token: %w", domain.ErrInvalidInput)
	}

	return s.repo.WithTx(ctx, func(u port.UnitOfWork) error {
		verification, err := u.Auth().GetVerifyEmailByID(ctx, id)
		if err != nil {
			return fmt.Errorf("get verification email: %w", err)
		}
		if verification.ID == uuid.Nil {
			return fmt.Errorf("verification token not found: %w", domain.ErrNotFound)
		}
		if time.Now().After(verification.ExpiredAt) {
			return fmt.Errorf("verification token expired: %w", domain.ErrInvalidInput)
		}
		if err := u.Auth().VerifyEmail(ctx, id); err != nil {
			return fmt.Errorf("verify email: %w", err)
		}
		return nil
	})
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	userFromDB, err := s.repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		return fmt.Errorf(errGetUserByEmail, err)
	}
	if userFromDB == nil {
		return fmt.Errorf(errUserNotFound, domain.ErrNotFound)
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
	hash, err := s.tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("hash refresh token: %w", err)
	}

	var accessToken, newRawToken string
	var reused bool
	err = s.repo.WithTx(ctx, func(u port.UnitOfWork) error {
		token, err := s.lookupRefreshToken(ctx, u.Auth(), hash)
		if err != nil {
			return err
		}
		if token.UsedAt != nil {
			reused = true
			if err := u.Auth().RevokeRefreshTokenFamily(ctx, token.FamilyID); err != nil {
				return fmt.Errorf("revoke token family: %w", err)
			}
			return nil
		}

		user, err := u.Users().GetByID(ctx, token.UserID)
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		if user == nil {
			return fmt.Errorf(errUserNotFound, domain.ErrNotFound)
		}
		accessToken, _, err = s.tokens.GenerateAccessToken(*user)
		if err != nil {
			return fmt.Errorf("generate access token: %w", err)
		}
		newRawToken, err = s.rotateRefreshToken(ctx, u.Auth(), token, ip, userAgent)
		return err
	})
	if err != nil {
		return "", "", err
	}
	if reused {
		return "", "", fmt.Errorf("%w: %w", ErrRefreshTokenReused, domain.ErrUnauthenticated)
	}
	return accessToken, newRawToken, nil
}

func (s *authService) lookupRefreshToken(ctx context.Context, auth port.AuthRepository, hash []byte) (*domain.RefreshToken, error) {
	token, err := auth.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	if token == nil {
		return nil, fmt.Errorf("refresh token not found: %w", domain.ErrUnauthenticated)
	}
	if time.Now().After(token.ExpiredAt) {
		return nil, fmt.Errorf("refresh token expired: %w", domain.ErrUnauthenticated)
	}
	if token.RevokedAt != nil {
		return nil, fmt.Errorf("refresh token revoked: %w", domain.ErrUnauthenticated)
	}
	return token, nil
}

func (s *authService) rotateRefreshToken(ctx context.Context, auth port.AuthRepository, token *domain.RefreshToken, ip, userAgent string) (string, error) {
	if err := auth.UseRefreshToken(ctx, token.ID); err != nil {
		return "", fmt.Errorf("use refresh token: %w", err)
	}

	raw, hmac, err := s.tokens.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	now := time.Now()
	newRT := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    token.UserID,
		TokenHash: hmac,
		FamilyID:  token.FamilyID,
		ParentID:  &token.ID,
		ExpiredAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		IPAddress: ip,
		UserAgent: userAgent,
	}
	if err := auth.CreateRefreshToken(ctx, newRT); err != nil {
		return "", fmt.Errorf("create new refresh token: %w", err)
	}
	return raw, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash, err := s.tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("hash refresh token: %w", err)
	}

	token, err := s.repo.Auth().GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("get refresh token: %w", err)
	}
	if token == nil {
		return nil
	}

	if err := s.repo.Auth().RevokeRefreshTokenFamily(ctx, token.FamilyID); err != nil {
		return fmt.Errorf("revoke token family: %w", err)
	}
	return nil
}
