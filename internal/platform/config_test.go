package platform

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{
		Port:                "8080",
		AllowedOrigins:      []string{"https://example.com"},
		Environment:         "development",
		BaseURL:             "https://api.example.com",
		DatabaseURL:         "postgres://example",
		JWTAccessSecret:     strings.Repeat("a", 32),
		RefreshTokenHashKey: strings.Repeat("b", 32),
		EmailAddress:        "auth@example.com",
		EmailPassword:       "password",
		SessionSecret:       strings.Repeat("c", 32),
		RateLimit: RateLimitConfig{
			Backend:       "memory",
			Key:           strings.Repeat("d", 32),
			MaxMemoryKeys: 100,
			Policies:      defaultRateLimitPolicies(),
		},
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	cfg.DatabaseURL = ""
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("missing database error = %v", err)
	}

	cfg = validConfig()
	cfg.GitHubClientID = "id"
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate succeeded with partial GitHub configuration")
	}

	cfg = validConfig()
	cfg.Environment = "production"
	cfg.JWTAccessSecret = "short"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "JWT_ACCESS_SECRET") {
		t.Fatalf("short production secret error = %v", err)
	}
}

func TestLoadConfigDefaultsAndOrigins(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("ALLOWED_ORIGINS", "https://one.example, https://two.example")
	cfg := LoadConfig()
	if cfg.Port != "8080" || cfg.Environment != "development" {
		t.Fatalf("defaults = port %q, environment %q", cfg.Port, cfg.Environment)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins = %#v", cfg.AllowedOrigins)
	}
}
