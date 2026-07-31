package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool   *pgxpool.Pool
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	ctx, cancel := context.WithCancel(context.Background())
	store := &PostgresStore{pool: pool, cancel: cancel}
	store.wg.Add(1)
	go store.cleanupLoop(ctx)
	return store
}

func (s *PostgresStore) Allow(ctx context.Context, key string, policy port.RateLimitPolicy) (port.RateLimitDecision, error) {
	if err := ctx.Err(); err != nil {
		return port.RateLimitDecision{}, err
	}
	if s.pool == nil || policy.Limit < 1 || policy.Window <= 0 {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}

	now := time.Now().UTC()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, key); err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	if _, err := tx.Exec(ctx, `DELETE FROM rate_limit_buckets WHERE key_hash = $1 AND expires_at <= $2`, key, now); err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}

	var count int
	var started time.Time
	err = tx.QueryRow(ctx, `
		SELECT count, window_start
		FROM rate_limit_buckets
		WHERE key_hash = $1
	`, key).Scan(&count, &started)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err := tx.Exec(ctx, `
			INSERT INTO rate_limit_buckets (key_hash, count, window_start, expires_at)
			VALUES ($1, 1, $2, $3)
		`, key, now, now.Add(policy.Window)); err != nil {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		if err := tx.Commit(ctx); err != nil {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		return port.RateLimitDecision{Allowed: true}, nil
	}
	if err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}

	resetAt := started.Add(policy.Window)
	if !now.Before(resetAt) {
		if _, err := tx.Exec(ctx, `
			UPDATE rate_limit_buckets
			SET count = 1, window_start = $2, expires_at = $3
			WHERE key_hash = $1
		`, key, now, now.Add(policy.Window)); err != nil {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		if err := tx.Commit(ctx); err != nil {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		return port.RateLimitDecision{Allowed: true}, nil
	}
	if count >= policy.Limit {
		if err := tx.Commit(ctx); err != nil {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		return port.RateLimitDecision{RetryAfter: resetAt.Sub(now)}, nil
	}

	if _, err := tx.Exec(ctx, `UPDATE rate_limit_buckets SET count = count + 1 WHERE key_hash = $1`, key); err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	if err := tx.Commit(ctx); err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	return port.RateLimitDecision{Allowed: true}, nil
}

func (s *PostgresStore) Reset(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.pool == nil {
		return port.ErrRateLimitBackendUnavailable
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM rate_limit_buckets WHERE key_hash = $1`, key); err != nil {
		return port.ErrRateLimitBackendUnavailable
	}
	return nil
}

func (s *PostgresStore) Backend() string { return "postgres" }

func (s *PostgresStore) Close() error {
	if s.cancel != nil {
		s.cancel()
		s.wg.Wait()
	}
	return nil
}

func (s *PostgresStore) cleanupLoop(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.pool != nil {
				_, _ = s.pool.Exec(ctx, `DELETE FROM rate_limit_buckets WHERE expires_at <= now()`)
			}
		}
	}
}

var _ port.RateLimitStore = (*PostgresStore)(nil)
