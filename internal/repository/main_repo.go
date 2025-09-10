// Package repository is the main repository package that provides access to all repositories and transaction management.
package repository

import (
	"context"
	"database/sql"
)

type Repository interface {
	Begin(ctx context.Context, opts *sql.TxOptions) (UOW, error)
	Auth() AuthRepository
	Users() UserRepository
	WithTx(ctx context.Context, fn func(u UOW) error) error
}

type UOW interface {
	Users() UserRepository
	Auth() AuthRepository
	Commit() error
	Rollback() error
}

type uow struct {
	tx *sql.Tx
}

type repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) Repository { return &repository{db: db} }

func (r *repository) Auth() AuthRepository  { return NewAuthRepository(r.db) }
func (r *repository) Users() UserRepository { return NewUserRepository(r.db) }

func (r *repository) Begin(ctx context.Context, opts *sql.TxOptions) (UOW, error) {
	tx, err := r.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &uow{tx: tx}, nil
}

func (u *uow) Users() UserRepository { return NewUserRepository(u.tx) }
func (u *uow) Auth() AuthRepository  { return NewAuthRepository(u.tx) }
func (u *uow) Commit() error         { return u.tx.Commit() }
func (u *uow) Rollback() error       { return u.tx.Rollback() }

func (r *repository) WithTx(ctx context.Context, fn func(u UOW) error) error {
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
