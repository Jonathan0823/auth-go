package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

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
	if err != nil {
		if pgUniqueViolation(err) {
			return fmt.Errorf("user already exists: %w", domain.ErrConflict)
		}
		return err
	}
	return nil
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

func (r *userRepository) Update(ctx context.Context, user domain.UpdateUserCommand) error {
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
