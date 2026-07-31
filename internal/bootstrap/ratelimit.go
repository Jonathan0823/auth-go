package bootstrap

import (
	"github.com/Jonathan0823/auth-go/internal/adapter/outbound/ratelimit"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func newRateLimitStore(config platform.RateLimitConfig, pool *pgxpool.Pool) (port.RateLimitStore, *redis.Client, error) {
	switch config.Backend {
	case "memory":
		return ratelimit.NewMemoryStore(config.MaxMemoryKeys), nil, nil
	case "postgres":
		return ratelimit.NewPostgresStore(pool), nil, nil
	case "redis":
		client := platform.NewRedisClient(config.RedisAddr, config.RedisPassword, config.RedisDB)
		return ratelimit.NewRedisStore(client), client, nil
	default:
		return nil, nil, port.ErrRateLimitBackendUnavailable
	}
}
