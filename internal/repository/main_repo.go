// Package repository is the main repository package that provides access to all repositories and transaction management.
package repository

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Auth() AuthRepository  { return NewAuthRepository(r.db) }
func (r *Repository) Users() UserRepository { return NewUserRepository(r.db) }

type UOW struct {
	tx *sql.Tx
}

func (r *Repository) Begin(ctx context.Context, opts *sql.TxOptions) (*UOW, error) {
	tx, err := r.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &UOW{tx: tx}, nil
}
func (u *UOW) Commit() error   { return u.tx.Commit() }
func (u *UOW) Rollback() error { return u.tx.Rollback() }

func (u *UOW) Users() UserRepository { return NewUserRepository(u.tx) }
func (u *UOW) Auth() AuthRepository  { return NewAuthRepository(u.tx) }

func (r *Repository) WithTx(ctx context.Context, fn func(u *UOW) error) error {
	u, err := r.Begin(ctx, nil)
	if err != nil {
		return err
	}
	defer u.Rollback()

	if err := fn(u); err != nil {
		return err
	}
	return u.Commit()
}
