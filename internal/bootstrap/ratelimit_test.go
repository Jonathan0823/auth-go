package bootstrap

import (
	"testing"

	"github.com/Jonathan0823/auth-go/internal/adapter/outbound/ratelimit"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
)

func TestNewRateLimitStoreBackends(t *testing.T) {
	memory, redisClient, err := newRateLimitStore(platform.RateLimitConfig{Backend: "memory", MaxMemoryKeys: 10}, nil)
	if err != nil || redisClient != nil || memory.Backend() != "memory" {
		t.Fatalf("memory store = %v, %v, %v", memory, redisClient, err)
	}
	_ = memory.Close()

	postgres, redisClient, err := newRateLimitStore(platform.RateLimitConfig{Backend: "postgres"}, nil)
	if err != nil || redisClient != nil || postgres.Backend() != "postgres" {
		t.Fatalf("postgres store = %v, %v, %v", postgres, redisClient, err)
	}
	_ = postgres.Close()

	redis, redisClient, err := newRateLimitStore(platform.RateLimitConfig{Backend: "redis", RedisAddr: "localhost:6399"}, nil)
	if err != nil || redisClient == nil || redis.Backend() != "redis" {
		t.Fatalf("redis store = %v, %v, %v", redis, redisClient, err)
	}
	_ = redis.Close()
	_ = redisClient.Close()

	store, client, err := newRateLimitStore(platform.RateLimitConfig{Backend: "unknown"}, nil)
	if store != nil || client != nil || err != port.ErrRateLimitBackendUnavailable {
		t.Fatalf("unknown backend = %v, %v, %v", store, client, err)
	}
}

var _ port.RateLimitStore = (*ratelimit.MemoryStore)(nil)
