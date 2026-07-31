package platform

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type recordingRateLimitStore struct {
	keys    []string
	backend string
}

func (s *recordingRateLimitStore) Allow(_ context.Context, key string, _ port.RateLimitPolicy) (port.RateLimitDecision, error) {
	s.keys = append(s.keys, key)
	return port.RateLimitDecision{Allowed: true}, nil
}
func (s *recordingRateLimitStore) Reset(_ context.Context, key string) error {
	s.keys = append(s.keys, key)
	return nil
}
func (s *recordingRateLimitStore) Backend() string { return s.backend }
func (s *recordingRateLimitStore) Close() error    { return nil }

func TestRateLimiterHMACProtectsKeys(t *testing.T) {
	store := &recordingRateLimitStore{backend: "memory"}
	limiter := NewRateLimiter(store, "secret", map[string]port.RateLimitPolicy{
		"login": {Name: "login", Limit: 1, Window: time.Minute},
	})
	value := "user@example.com"
	if _, err := limiter.Allow(context.Background(), RateLimitCheck{
		Policy: "login", Dimension: "email", Value: value,
	}); err != nil {
		t.Fatal(err)
	}
	if len(store.keys) != 1 || strings.Contains(store.keys[0], value) {
		t.Fatalf("stored key = %q; raw value leaked", store.keys)
	}
}
