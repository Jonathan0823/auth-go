package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/core/ratelimit"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type RateLimitError struct {
	Policy      string
	Backend     string
	RetryAfter  time.Duration
	Unavailable bool
}

func (e *RateLimitError) Error() string {
	if e.Unavailable {
		return "rate limit backend unavailable"
	}
	return "rate limit exceeded"
}

func (e *RateLimitError) Unwrap() error {
	if e.Unavailable {
		return port.ErrRateLimitBackendUnavailable
	}
	return ErrRateLimitExceeded
}

func (s *authService) allowRateLimit(ctx context.Context, policy, dimension, value string) error {
	check := ratelimit.RateLimitCheck{Policy: policy, Dimension: dimension, Value: value}
	if s.rateLimiter == nil {
		return &RateLimitError{Policy: policy, Backend: "unknown", Unavailable: true}
	}
	decision, err := s.rateLimiter.Allow(ctx, check)
	if err != nil {
		return &RateLimitError{Policy: policy, Backend: s.rateLimiter.Backend(), Unavailable: true}
	}
	if !decision.Allowed {
		return &RateLimitError{
			Policy:     policy,
			Backend:    s.rateLimiter.Backend(),
			RetryAfter: decision.RetryAfter,
		}
	}
	return nil
}

func (s *authService) resetRateLimit(ctx context.Context, policy, dimension, value string) {
	if s.rateLimiter == nil {
		return
	}
	_ = s.rateLimiter.Reset(ctx, ratelimit.RateLimitCheck{Policy: policy, Dimension: dimension, Value: value})
}

func normalizeRateLimitEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
