//go:build integration

package ratelimit

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresStoreContract(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := NewPostgresStore(pool)
	defer store.Close()
	runStoreContract(t, store)
}

func TestPostgresStoreConcurrentLimit(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := NewPostgresStore(pool)
	defer store.Close()
	policy := port.RateLimitPolicy{Name: "concurrent", Limit: 5, Window: time.Minute}
	key := "concurrent-" + time.Now().Format("150405.000000000")
	var wg sync.WaitGroup
	results := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			decision, err := store.Allow(context.Background(), key, policy)
			results <- err == nil && decision.Allowed
		}()
	}
	wg.Wait()
	close(results)

	allowed := 0
	for result := range results {
		if result {
			allowed++
		}
	}
	if allowed != policy.Limit {
		t.Fatalf("allowed = %d, want %d", allowed, policy.Limit)
	}
	if err := store.Reset(context.Background(), key); err != nil {
		t.Fatal(err)
	}
}
