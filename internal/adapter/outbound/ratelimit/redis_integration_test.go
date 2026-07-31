//go:build integration

package ratelimit

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func newIntegrationRedisStore(t *testing.T) *RedisStore {
	t.Helper()
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		t.Skip("REDIS_ADDR is not set")
	}
	database := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		var err error
		database, err = strconv.Atoi(raw)
		if err != nil {
			t.Fatal(err)
		}
	}
	client := platform.NewRedisClient(redisAddr, os.Getenv("REDIS_PASSWORD"), database)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return NewRedisStore(client)
}

func TestRedisStoreContract(t *testing.T) {
	store := newIntegrationRedisStore(t)
	runStoreContract(t, store)
}

func TestRedisStoreConcurrentLimit(t *testing.T) {
	store := newIntegrationRedisStore(t)
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
