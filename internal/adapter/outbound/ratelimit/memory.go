package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type MemoryStore struct {
	mu      sync.Mutex
	entries map[string]memoryEntry
	maxKeys int
	now     func() time.Time
}

type memoryEntry struct {
	count   int
	started time.Time
	window  time.Duration
}

func NewMemoryStore(maxKeys int) *MemoryStore {
	if maxKeys < 1 {
		maxKeys = 10_000
	}
	return &MemoryStore{
		entries: make(map[string]memoryEntry),
		maxKeys: maxKeys,
		now:     time.Now,
	}
}

func (s *MemoryStore) Allow(ctx context.Context, key string, policy port.RateLimitPolicy) (port.RateLimitDecision, error) {
	if err := ctx.Err(); err != nil {
		return port.RateLimitDecision{}, err
	}
	if policy.Limit < 1 || policy.Window <= 0 {
		return port.RateLimitDecision{}, port.ErrRateLimitBackendUnavailable
	}

	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()

	for storedKey, entry := range s.entries {
		if !now.Before(entry.started.Add(entry.window)) {
			delete(s.entries, storedKey)
		}
	}
	entry, exists := s.entries[key]
	if !exists || !now.Before(entry.started.Add(entry.window)) {
		if !exists && len(s.entries) >= s.maxKeys {
			s.evictOldest()
		}
		s.entries[key] = memoryEntry{count: 1, started: now, window: policy.Window}
		return port.RateLimitDecision{Allowed: true}, nil
	}

	resetAt := entry.started.Add(entry.window)
	if entry.count >= policy.Limit {
		return port.RateLimitDecision{
			RetryAfter: resetAt.Sub(now),
		}, nil
	}
	entry.count++
	s.entries[key] = entry
	return port.RateLimitDecision{Allowed: true}, nil
}

func (s *MemoryStore) Reset(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.entries, key)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) Backend() string { return "memory" }

func (s *MemoryStore) Close() error { return nil }

func (s *MemoryStore) evictOldest() {
	var oldestKey string
	var oldest time.Time
	for key, entry := range s.entries {
		if oldestKey == "" || entry.started.Before(oldest) {
			oldestKey = key
			oldest = entry.started
		}
	}
	if oldestKey != "" {
		delete(s.entries, oldestKey)
	}
}
