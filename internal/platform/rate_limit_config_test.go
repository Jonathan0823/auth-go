package platform

import (
	"testing"
)

func TestLoadRateLimitConfigDefaults(t *testing.T) {
	for _, name := range []string{
		"RATE_LIMIT_BACKEND", "RATE_LIMIT_KEY", "TRUSTED_PROXIES",
		"RATE_LIMIT_ALLOW_MEMORY_PRODUCTION", "RATE_LIMIT_MAX_MEMORY_KEYS",
	} {
		t.Setenv(name, "")
	}
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("RATE_LIMIT_KEY", "test-key")
	config := LoadRateLimitConfig()
	if config.Backend != "memory" || config.MaxMemoryKeys != 10_000 {
		t.Fatalf("defaults = %#v", config)
	}
	if err := config.Validate("development"); err != nil {
		t.Fatalf("development validation = %v", err)
	}
	if err := config.Validate("production"); err == nil {
		t.Fatal("memory backend should be rejected in production")
	}
}

func TestRateLimitConfigValidatesTrustedProxiesAndPolicies(t *testing.T) {
	t.Setenv("RATE_LIMIT_BACKEND", "redis")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "0")
	t.Setenv("RATE_LIMIT_KEY", "test-key")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 192.168.1.1")
	t.Setenv("RATE_LIMIT_LOGIN_IP", "2/10s")
	config := LoadRateLimitConfig()
	if err := config.Validate("production"); err != nil {
		t.Fatal(err)
	}
	if got := config.Policies["login_ip"]; got.Limit != 2 || got.Window.String() != "10s" {
		t.Fatalf("login policy = %#v", got)
	}

	t.Setenv("TRUSTED_PROXIES", "not-a-cidr")
	if err := LoadRateLimitConfig().Validate("production"); err == nil {
		t.Fatal("invalid trusted proxy should fail validation")
	}
}
