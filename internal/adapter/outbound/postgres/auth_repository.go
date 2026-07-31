package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *authRepository) CreateRefreshToken(ctx context.Context, rt domain.RefreshToken) error {
	return r.q.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		ID:        pgtypeUUID(rt.ID),
		UserID:    int32(rt.UserID),
		TokenHash: rt.TokenHash,
		FamilyID:  pgtypeUUID(rt.FamilyID),
		ParentID:  pgtypeUUIDPtr(rt.ParentID),
		ExpiredAt: pgtypeTimestamp(rt.ExpiredAt),
		CreatedAt: pgtypeTimestamp(rt.CreatedAt),
		IpAddress: rt.IPAddress,
		UserAgent: rt.UserAgent,
	})
}

func (r *authRepository) GetRefreshTokenByHash(ctx context.Context, hash []byte) (*domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.RefreshToken{
		ID:        row.ID.Bytes,
		UserID:    int(row.UserID),
		TokenHash: row.TokenHash,
		FamilyID:  row.FamilyID.Bytes,
		ParentID:  uuidPtr(row.ParentID),
		ExpiredAt: row.ExpiredAt.Time,
		UsedAt:    timestampPtr(row.UsedAt),
		RevokedAt: timestampPtr(row.RevokedAt),
		CreatedAt: row.CreatedAt.Time,
		IPAddress: row.IpAddress,
		UserAgent: row.UserAgent,
	}, nil
}

func (r *authRepository) UseRefreshToken(ctx context.Context, id uuid.UUID) error {
	return r.q.UseRefreshToken(ctx, pgtypeUUID(id))
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	return r.q.RevokeRefreshToken(ctx, pgtypeUUID(id))
}

func (r *authRepository) RevokeRefreshTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	return r.q.RevokeRefreshTokenFamily(ctx, pgtypeUUID(familyID))
}

func pgtypeUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func uuidPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	u := uuid.UUID(id.Bytes)
	return &u
}

func timestampPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
