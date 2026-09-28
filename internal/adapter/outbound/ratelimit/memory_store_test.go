package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

func TestMemoryStoreContract(t *testing.T) {
	runStoreContract(t, NewMemoryStore(10))
}

func TestMemoryStoreEnforcesAndResetsWindow(t *testing.T) {
	now := time.Unix(0, 0)
	store := NewMemoryStore(10)
	store.now = func() time.Time { return now }
	policy := port.RateLimitPolicy{Name: "test", Limit: 2, Window: time.Minute}

	for i := 0; i < 2; i++ {
		decision, err := store.Allow(context.Background(), "key", policy)
		if err != nil || !decision.Allowed {
			t.Fatalf("request %d = %#v, %v; want allowed", i, decision, err)
		}
	}
	decision, err := store.Allow(context.Background(), "key", policy)
	if err != nil || decision.Allowed || decision.RetryAfter != time.Minute {
		t.Fatalf("limited request = %#v, %v; want one minute retry", decision, err)
	}

	now = now.Add(time.Minute)
	decision, err = store.Allow(context.Background(), "key", policy)
	if err != nil || !decision.Allowed {
		t.Fatalf("request after expiry = %#v, %v; want allowed", decision, err)
	}
	if err := store.Reset(context.Background(), "key"); err != nil {
		t.Fatal(err)
	}
}

func TestMemoryStoreBoundsKeys(t *testing.T) {
	store := NewMemoryStore(2)
	policy := port.RateLimitPolicy{Name: "test", Limit: 1, Window: time.Minute}
	for _, key := range []string{"one", "two", "three"} {
		if _, err := store.Allow(context.Background(), key, policy); err != nil {
			t.Fatal(err)
		}
	}
	if len(store.entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(store.entries))
	}
}

func TestMemoryStoreConcurrentAccess(t *testing.T) {
	store := NewMemoryStore(100)
	policy := port.RateLimitPolicy{Name: "test", Limit: 100, Window: time.Minute}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.Allow(context.Background(), "key", policy)
		}()
	}
	wg.Wait()
}
