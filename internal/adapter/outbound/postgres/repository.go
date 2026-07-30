package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

// Ensure adapter implements port interfaces.
var _ port.Repository = (*repository)(nil)
var _ port.UnitOfWork = (*unitOfWork)(nil)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) port.Repository {
	return &repository{pool: pool}
}

func (r *repository) Users() port.UserRepository {
	q := New(r.pool)
	return &userRepository{q: q, pool: r.pool}
}

func (r *repository) Auth() port.AuthRepository {
	q := New(r.pool)
	return &authRepository{q: q, pool: r.pool}
}

func (r *repository) Begin(ctx context.Context) (port.UnitOfWork, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	return &unitOfWork{tx: tx, pool: r.pool}, nil
}

func (r *repository) WithTx(ctx context.Context, fn func(u port.UnitOfWork) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	u := &unitOfWork{tx: tx, pool: r.pool}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(u); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type unitOfWork struct {
	tx   pgx.Tx
	pool *pgxpool.Pool
}

func (u *unitOfWork) Users() port.UserRepository {
	return &userRepository{q: New(u.tx), pool: u.pool}
}

func (u *unitOfWork) Auth() port.AuthRepository {
	return &authRepository{q: New(u.tx), pool: u.pool}
}

func (u *unitOfWork) Commit() error   { return u.tx.Commit(context.Background()) }
func (u *unitOfWork) Rollback() error { return u.tx.Rollback(context.Background()) }

type userRepository struct {
	q    *Queries
	pool *pgxpool.Pool
}

func pgUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *userRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.User{
		ID:         int(row.ID),
		Username:   row.Username,
		Email:      row.Email,
		IsVerified: row.IsVerified.Bool,
		UpdatedAt:  row.UpdatedAt.Time,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string, includePassword bool) (*domain.User, error) {
	if includePassword {
		row, err := r.q.GetUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return &domain.User{
			ID:         int(row.ID),
			Username:   row.Username,
			Email:      row.Email,
			Password:   row.Password,
			IsVerified: row.IsVerified.Bool,
			UpdatedAt:  row.UpdatedAt.Time,
			CreatedAt:  row.CreatedAt.Time,
		}, nil
	}
	row, err := r.q.GetUserByEmailWithoutPassword(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.User{
		ID:         int(row.ID),
		Username:   row.Username,
		Email:      row.Email,
		IsVerified: row.IsVerified.Bool,
		UpdatedAt:  row.UpdatedAt.Time,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

func (r *userRepository) Create(ctx context.Context, user domain.User) error {
	_, err := r.q.CreateUser(ctx, CreateUserParams{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	})
	return err
}

func (r *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.q.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, &domain.User{
			ID:         int(row.ID),
			Username:   row.Username,
			Email:      row.Email,
			IsVerified: row.IsVerified.Bool,
			UpdatedAt:  row.UpdatedAt.Time,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user domain.UpdateUserRequest) error {
	return r.q.UpdateUser(ctx, UpdateUserParams{
		Username: user.Username,
		Email:    user.Email,
		ID:       int32(user.ID),
	})
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	return r.q.DeleteUser(ctx, int32(id))
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	return r.q.UpdateUserPassword(ctx, UpdateUserPasswordParams{
		Password: newPassword,
		ID:       int32(id),
	})
}

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
