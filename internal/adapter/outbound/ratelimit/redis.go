package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/redis/go-redis/v9"
)

var consumeScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if not current then
  redis.call("SET", KEYS[1], 1, "PX", ARGV[2])
  return {1, ARGV[2]}
end
if tonumber(current) >= tonumber(ARGV[1]) then
  return {0, redis.call("PTTL", KEYS[1])}
end
redis.call("INCR", KEYS[1])
return {1, redis.call("PTTL", KEYS[1])}
`)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (s *RedisStore) Allow(ctx context.Context, key string, policy port.RateLimitPolicy) (port.RateLimitDecision, error) {
	if err := ctx.Err(); err != nil {
		return port.RateLimitDecision{}, err
	}
	if s.client == nil || policy.Limit < 1 || policy.Window <= 0 {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	window := policy.Window.Milliseconds()
	if window < 1 {
		window = 1
	}
	result, err := consumeScript.Run(ctx, s.client, []string{key}, policy.Limit, window).Result()
	if err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	values, ok := result.([]any)
	if !ok || len(values) != 2 {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	allowed, err := strconv.ParseInt(toString(values[0]), 10, 64)
	if err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	ttl, err := strconv.ParseInt(toString(values[1]), 10, 64)
	if err != nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	decision := port.RateLimitDecision{Allowed: allowed == 1}
	if !decision.Allowed {
		decision.RetryAfter = time.Duration(ttl) * time.Millisecond
	}
	return decision, nil
}

func (s *RedisStore) Reset(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.client == nil {
		return port.ErrRateLimitBackendUnavailable
	}
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return port.ErrRateLimitBackendUnavailable
	}
	return nil
}

func (s *RedisStore) Backend() string { return "redis" }

func (s *RedisStore) Close() error { return nil }

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

var _ port.RateLimitStore = (*RedisStore)(nil)
