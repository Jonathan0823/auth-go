package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

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
	return &userRepository{q: New(r.pool)}
}

func (r *repository) Auth() port.AuthRepository {
	return &authRepository{q: New(r.pool)}
}

func (r *repository) WithTx(ctx context.Context, fn func(port.UnitOfWork) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(&unitOfWork{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type unitOfWork struct {
	tx pgx.Tx
}

func (u *unitOfWork) Users() port.UserRepository {
	return &userRepository{q: New(u.tx)}
}

func (u *unitOfWork) Auth() port.AuthRepository {
	return &authRepository{q: New(u.tx)}
}
