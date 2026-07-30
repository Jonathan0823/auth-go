package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

type authRepository struct {
	q    *Queries
	pool *pgxpool.Pool
}

func (r *authRepository) CreateVerifyEmail(ctx context.Context, ve domain.VerifyEmail) error {
	return r.q.CreateVerifyEmail(ctx, CreateVerifyEmailParams{
		ID:        pgtypeUUID(ve.ID),
		UserID:    int32(ve.UserID),
		Email:     ve.Email,
		ExpiredAt: pgtypeTimestamp(ve.ExpiredAt),
	})
}

func (r *authRepository) GetVerifyEmailByID(ctx context.Context, id string) (domain.VerifyEmail, error) {
	uid, err := pgtypeUUIDFromString(id)
	if err != nil {
		return domain.VerifyEmail{}, err
	}
	row, err := r.q.GetVerifyEmailByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VerifyEmail{}, nil
		}
		return domain.VerifyEmail{}, err
	}
	return domain.VerifyEmail{
		ID:        row.ID.Bytes,
		Email:     row.Email,
		ExpiredAt: row.ExpiredAt.Time,
	}, nil
}

func (r *authRepository) VerifyEmail(ctx context.Context, id string) error {
	uid, err := pgtypeUUIDFromString(id)
	if err != nil {
		return err
	}
	if err := r.q.VerifyEmailByToken(ctx, uid); err != nil {
		return err
	}
	return r.q.VerifyEmailDeleteToken(ctx, uid)
}

func (r *authRepository) CreateForgotPasswordEmail(ctx context.Context, data domain.ForgotPassword) error {
	return r.q.CreateForgotPasswordEmail(ctx, CreateForgotPasswordEmailParams{
		ID:        pgtypeUUID(data.ID),
		UserID:    int32(data.UserID),
		Email:     data.Email,
		ExpiredAt: pgtypeTimestamp(data.ExpiredAt),
	})
}

func (r *authRepository) GetForgotPasswordByID(ctx context.Context, id string) (domain.ForgotPassword, error) {
	uid, err := pgtypeUUIDFromString(id)
	if err != nil {
		return domain.ForgotPassword{}, err
	}
	row, err := r.q.GetForgotPasswordByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ForgotPassword{}, nil
		}
		return domain.ForgotPassword{}, err
	}
	return domain.ForgotPassword{
		ID:        row.ID.Bytes,
		UserID:    int(row.UserID),
		Email:     row.Email,
		ExpiredAt: row.ExpiredAt.Time,
	}, nil
}

func (r *authRepository) DeleteForgotPasswordByID(ctx context.Context, id string) error {
	uid, err := pgtypeUUIDFromString(id)
	if err != nil {
		return err
	}
	return r.q.DeleteForgotPasswordByID(ctx, uid)
}

func (r *authRepository) CreateTokenLog(ctx context.Context, tl domain.TokenLog) error {
	return r.q.CreateTokenLog(ctx, CreateTokenLogParams{
		ID:               pgtypeUUID(tl.ID),
		UserID:           int32(tl.UserID),
		Jti:              tl.JTI,
		RefreshedFromJti: pgtypeTextPtr(tl.RefreshedFromJTI),
		InvalidatedAt:    pgtypeTimestampPtr(tl.InvalidatedAt),
		ExpiredAt:        pgtypeTimestamp(tl.ExpiredAt),
		CreatedAt:        pgtypeTimestamp(tl.CreatedAt),
		IpAddress:        tl.IPAddress,
		UserAgent:        tl.UserAgent,
	})
}

func (r *authRepository) GetTokenLogByJTI(ctx context.Context, jti string) (domain.TokenLog, error) {
	row, err := r.q.GetTokenLogByJTI(ctx, jti)
	if err != nil {
		return domain.TokenLog{}, err
	}
	return domain.TokenLog{
		ID:               row.ID.Bytes,
		UserID:           int(row.UserID),
		JTI:              row.Jti,
		RefreshedFromJTI: textPtr(row.RefreshedFromJti),
		ExpiredAt:        row.ExpiredAt.Time,
		CreatedAt:        row.CreatedAt.Time,
		IPAddress:        row.IpAddress,
		UserAgent:        row.UserAgent,
	}, nil
}

func (r *authRepository) InvalidateTokenLog(ctx context.Context, oldJTI, newJTI string) error {
	if newJTI == "" {
		return r.q.InvalidateTokenLog(ctx, oldJTI)
	}
	return r.q.InvalidateAndRefreshTokenLog(ctx, InvalidateAndRefreshTokenLogParams{
		Jti:              oldJTI,
		RefreshedFromJti: pgtypeText(newJTI),
	})
}

func (r *authRepository) IsTokenLogInvalidated(ctx context.Context, jti string) (bool, error) {
	return r.q.IsTokenLogInvalidated(ctx, jti)
}
