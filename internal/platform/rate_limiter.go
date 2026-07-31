package platform

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type RateLimitCheck struct {
	Policy    string
	Dimension string
	Value     string
}

type RateLimiter struct {
	store    port.RateLimitStore
	key      []byte
	policies map[string]port.RateLimitPolicy
}

func NewRateLimiter(store port.RateLimitStore, key string, policies map[string]port.RateLimitPolicy) *RateLimiter {
	return &RateLimiter{store: store, key: []byte(key), policies: policies}
}

func (l *RateLimiter) Allow(ctx context.Context, checks ...RateLimitCheck) (port.RateLimitDecision, error) {
	if l == nil || l.store == nil {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}
	var retryAfter time.Duration
	for _, check := range checks {
		policy, ok := l.policies[check.Policy]
		if !ok || policy.Limit < 1 || policy.Window <= 0 {
			return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
		}
		decision, err := l.store.Allow(ctx, l.keyFor(check), policy)
		if err != nil {
			return port.RateLimitDecision{}, err
		}
		if !decision.Allowed && decision.RetryAfter > retryAfter {
			retryAfter = decision.RetryAfter
		}
		if !decision.Allowed {
			return port.RateLimitDecision{RetryAfter: retryAfter}, nil
		}
	}
	return port.RateLimitDecision{Allowed: true}, nil
}

func (l *RateLimiter) Reset(ctx context.Context, check RateLimitCheck) error {
	if l == nil || l.store == nil {
		return port.ErrRateLimitBackendUnavailable
	}
	if _, ok := l.policies[check.Policy]; !ok {
		return port.ErrRateLimitBackendUnavailable
	}
	return l.store.Reset(ctx, l.keyFor(check))
}

func (l *RateLimiter) Backend() string {
	if l == nil || l.store == nil {
		return "unknown"
	}
	return l.store.Backend()
}

func (l *RateLimiter) keyFor(check RateLimitCheck) string {
	mac := hmac.New(sha256.New, l.key)
	_, _ = mac.Write([]byte(check.Policy))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(check.Dimension))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(check.Value))
	return hex.EncodeToString(mac.Sum(nil))
}
