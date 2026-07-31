package platform

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type RateLimitConfig struct {
	Backend                    string
	RedisAddr                  string
	RedisPassword              string
	RedisDB                    int
	Key                        string
	TrustedProxies             []string
	AllowMemoryInProduction    bool
	MaxMemoryKeys              int
	Policies                   map[string]port.RateLimitPolicy
	invalidPolicyConfiguration []string
}

func LoadRateLimitConfig() RateLimitConfig {
	backend := os.Getenv("RATE_LIMIT_BACKEND")
	if backend == "" {
		backend = "memory"
	}
	maxKeys := 10_000
	invalidMaxKeys := false
	if raw := os.Getenv("RATE_LIMIT_MAX_MEMORY_KEYS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			invalidMaxKeys = true
		} else {
			maxKeys = parsed
		}
	}
	config := RateLimitConfig{
		Backend:                 backend,
		RedisAddr:               os.Getenv("REDIS_ADDR"),
		RedisPassword:           os.Getenv("REDIS_PASSWORD"),
		Key:                     os.Getenv("RATE_LIMIT_KEY"),
		TrustedProxies:          splitList(os.Getenv("TRUSTED_PROXIES")),
		AllowMemoryInProduction: os.Getenv("RATE_LIMIT_ALLOW_MEMORY_PRODUCTION") == "true",
		MaxMemoryKeys:           maxKeys,
		Policies:                make(map[string]port.RateLimitPolicy),
	}
	if invalidMaxKeys {
		config.invalidPolicyConfiguration = append(config.invalidPolicyConfiguration, "RATE_LIMIT_MAX_MEMORY_KEYS")
	}
	config.RedisDB = loadRedisDB(&config)
	for name, fallback := range defaultRateLimitPolicies() {
		config.Policies[name] = loadRateLimitPolicy(&config, name, fallback)
	}
	return config
}

func (c RateLimitConfig) Validate(environment string) error {
	switch c.Backend {
	case "memory":
		if environment == "production" && !c.AllowMemoryInProduction {
			return fmt.Errorf("memory rate-limit backend is disabled in production")
		}
	case "redis":
		if c.RedisAddr == "" {
			return fmt.Errorf("REDIS_ADDR is required for the redis rate-limit backend")
		}
		if c.RedisDB < 0 {
			return fmt.Errorf("REDIS_DB must not be negative")
		}
	case "postgres":
	default:
		return fmt.Errorf("unsupported rate-limit backend %q", c.Backend)
	}
	if c.Key == "" {
		return fmt.Errorf("RATE_LIMIT_KEY is required")
	}
	if c.MaxMemoryKeys < 1 {
		return fmt.Errorf("RATE_LIMIT_MAX_MEMORY_KEYS must be positive")
	}
	if len(c.invalidPolicyConfiguration) > 0 {
		return fmt.Errorf("invalid rate-limit policies: %s", strings.Join(c.invalidPolicyConfiguration, ", "))
	}
	for _, proxy := range c.TrustedProxies {
		if net.ParseIP(proxy) == nil {
			if _, _, err := net.ParseCIDR(proxy); err != nil {
				return fmt.Errorf("invalid trusted proxy address %q", proxy)
			}
		}
	}
	return nil
}

func defaultRateLimitPolicies() map[string]port.RateLimitPolicy {
	return map[string]port.RateLimitPolicy{
		"login_ip":       {Name: "login_ip", Limit: 30, Window: time.Minute},
		"login_account":  {Name: "login_account", Limit: 5, Window: 15 * time.Minute},
		"register_ip":    {Name: "register_ip", Limit: 5, Window: time.Hour},
		"register_email": {Name: "register_email", Limit: 5, Window: time.Hour},
		"recovery_ip":    {Name: "recovery_ip", Limit: 10, Window: time.Hour},
		"recovery_email": {Name: "recovery_email", Limit: 3, Window: time.Hour},
		"recovery_token": {Name: "recovery_token", Limit: 3, Window: time.Hour},
		"refresh_ip":     {Name: "refresh_ip", Limit: 30, Window: time.Minute},
		"verify_ip":      {Name: "verify_ip", Limit: 10, Window: time.Hour},
		"verify_email":   {Name: "verify_email", Limit: 3, Window: time.Hour},
		"oauth_ip":       {Name: "oauth_ip", Limit: 10, Window: time.Minute},
	}
}

func loadRateLimitPolicy(config *RateLimitConfig, name string, fallback port.RateLimitPolicy) port.RateLimitPolicy {
	envName := "RATE_LIMIT_" + strings.ToUpper(name)
	raw := os.Getenv(envName)
	if raw == "" {
		return fallback
	}
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 {
		config.invalidPolicyConfiguration = append(config.invalidPolicyConfiguration, envName)
		return port.RateLimitPolicy{Name: name}
	}
	limit, limitErr := strconv.Atoi(parts[0])
	window, windowErr := time.ParseDuration(parts[1])
	if limitErr != nil || windowErr != nil || limit < 1 || window <= 0 {
		config.invalidPolicyConfiguration = append(config.invalidPolicyConfiguration, envName)
		return port.RateLimitPolicy{Name: name}
	}
	return port.RateLimitPolicy{Name: name, Limit: limit, Window: window}
}

func loadRedisDB(config *RateLimitConfig) int {
	raw := os.Getenv("REDIS_DB")
	if raw == "" {
		return 0
	}
	db, err := strconv.Atoi(raw)
	if err != nil || db < 0 {
		config.invalidPolicyConfiguration = append(config.invalidPolicyConfiguration, "REDIS_DB")
		return -1
	}
	return db
}

func splitList(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			result = append(result, item)
		}
	}
	return result
}
