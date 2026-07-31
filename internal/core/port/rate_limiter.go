package port

import (
	"context"
	"errors"
	"time"
)

var ErrRateLimitBackendUnavailable = errors.New("rate limit backend unavailable")

type RateLimitPolicy struct {
	Name   string
	Limit  int
	Window time.Duration
}

type RateLimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}

type RateLimitStore interface {
	Allow(context.Context, string, RateLimitPolicy) (RateLimitDecision, error)
	Reset(context.Context, string) error
	Backend() string
	Close() error
}
