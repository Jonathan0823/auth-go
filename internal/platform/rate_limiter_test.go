package platform

import (
	"context"
	"errors"
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

type decisionRateLimitStore struct {
	decisions []port.RateLimitDecision
	err       error
	resetErr  error
	calls     int
}

func (s *decisionRateLimitStore) Allow(context.Context, string, port.RateLimitPolicy) (port.RateLimitDecision, error) {
	if s.err != nil {
		return port.RateLimitDecision{}, s.err
	}
	decision := port.RateLimitDecision{Allowed: true}
	if s.calls < len(s.decisions) {
		decision = s.decisions[s.calls]
	}
	s.calls++
	return decision, nil
}
func (s *decisionRateLimitStore) Reset(context.Context, string) error { return s.resetErr }
func (s *decisionRateLimitStore) Backend() string                     { return "memory" }
func (s *decisionRateLimitStore) Close() error                        { return nil }

func TestRateLimiterRejectsInvalidChecksAndBackendFailures(t *testing.T) {
	valid := map[string]port.RateLimitPolicy{"login": {Name: "login", Limit: 1, Window: time.Minute}}
	for name, limiter := range map[string]*RateLimiter{
		"nil":       nil,
		"nil store": NewRateLimiter(nil, "key", valid),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := limiter.Allow(context.Background(), RateLimitCheck{Policy: "login"}); !errors.Is(err, port.ErrRateLimitBackendUnavailable) {
				t.Fatalf("Allow error = %v", err)
			}
		})
	}

	for _, check := range []RateLimitCheck{{Policy: "missing"}, {Policy: "login"}} {
		store := &decisionRateLimitStore{}
		policies := valid
		if check.Policy == "login" {
			policies = map[string]port.RateLimitPolicy{"login": {Name: "login", Limit: 0, Window: time.Minute}}
		}
		if _, err := NewRateLimiter(store, "key", policies).Allow(context.Background(), check); !errors.Is(err, port.ErrRateLimitBackendUnavailable) {
			t.Fatalf("invalid check error = %v", err)
		}
	}

	store := &decisionRateLimitStore{err: errors.New("backend failure")}
	if _, err := NewRateLimiter(store, "key", valid).Allow(context.Background(), RateLimitCheck{Policy: "login"}); !errors.Is(err, store.err) {
		t.Fatalf("backend error = %v", err)
	}
}

func TestRateLimiterShortCircuitsDeniedChecks(t *testing.T) {
	store := &decisionRateLimitStore{decisions: []port.RateLimitDecision{
		{Allowed: true},
		{Allowed: false, RetryAfter: 10 * time.Second},
	}}
	limiter := NewRateLimiter(store, "key", map[string]port.RateLimitPolicy{
		"login": {Name: "login", Limit: 1, Window: time.Minute},
	})
	decision, err := limiter.Allow(context.Background(),
		RateLimitCheck{Policy: "login", Dimension: "ip", Value: "one"},
		RateLimitCheck{Policy: "login", Dimension: "email", Value: "two"},
	)
	if err != nil || decision.Allowed || decision.RetryAfter != 10*time.Second || store.calls != 2 {
		t.Fatalf("decision = %#v, err=%v, calls=%d", decision, err, store.calls)
	}
}

func TestRateLimiterResetAndContext(t *testing.T) {
	store := &decisionRateLimitStore{}
	limiter := NewRateLimiter(store, "key", map[string]port.RateLimitPolicy{
		"login": {Name: "login", Limit: 1, Window: time.Minute},
	})
	if err := limiter.Reset(context.Background(), RateLimitCheck{Policy: "missing"}); !errors.Is(err, port.ErrRateLimitBackendUnavailable) {
		t.Fatalf("unknown reset error = %v", err)
	}
	store.resetErr = errors.New("reset failure")
	if err := limiter.Reset(context.Background(), RateLimitCheck{Policy: "login", Value: "user"}); !errors.Is(err, store.resetErr) {
		t.Fatalf("reset error = %v", err)
	}
	if _, err := limiter.Allow(context.Background(), RateLimitCheck{Policy: "login"}); err != nil {
		t.Fatalf("valid Allow = %v", err)
	}
}
