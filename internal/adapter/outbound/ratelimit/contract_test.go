package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

func runStoreContract(t *testing.T, store port.RateLimitStore) {
	t.Helper()
	ctx := context.Background()
	policy := port.RateLimitPolicy{Name: "contract", Limit: 2, Window: time.Minute}
	key := "contract-" + time.Now().Format("150405.000000000")

	for i := 0; i < 2; i++ {
		decision, err := store.Allow(ctx, key, policy)
		if err != nil || !decision.Allowed {
			t.Fatalf("request %d = %#v, %v; want allowed", i, decision, err)
		}
	}
	decision, err := store.Allow(ctx, key, policy)
	if err != nil || decision.Allowed || decision.RetryAfter <= 0 {
		t.Fatalf("limited request = %#v, %v; want retry duration", decision, err)
	}
	if err := store.Reset(ctx, key); err != nil {
		t.Fatal(err)
	}
	decision, err = store.Allow(ctx, key, policy)
	if err != nil || !decision.Allowed {
		t.Fatalf("request after reset = %#v, %v; want allowed", decision, err)
	}
}
