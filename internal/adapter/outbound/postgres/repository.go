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
	return &userRepository{q: New(r.pool), pool: r.pool}
}

func (r *repository) Auth() port.AuthRepository {
	return &authRepository{q: New(r.pool), pool: r.pool}
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
