package platform

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Port                string
	AllowedOrigins      []string
	Environment         string
	BaseURL             string
	LogLevel            string
	Domain              string
	DatabaseURL         string
	JWTAccessSecret     string
	RefreshTokenHashKey string
	EmailAddress        string
	EmailPassword       string
	SessionSecret       string
	GitHubClientID      string
	GitHubClientSecret  string
	GoogleClientID      string
	GoogleClientSecret  string
	EnableSwagger       bool
	EnableMetrics       bool
	RateLimit           RateLimitConfig
}

func LoadConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	return Config{
		Port:                port,
		AllowedOrigins:      splitList(os.Getenv("ALLOWED_ORIGINS")),
		Environment:         environment,
		BaseURL:             os.Getenv("BASE_URL"),
		LogLevel:            os.Getenv("LOG_LEVEL"),
		Domain:              os.Getenv("DOMAIN"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTAccessSecret:     os.Getenv("JWT_ACCESS_SECRET"),
		RefreshTokenHashKey: os.Getenv("REFRESH_TOKEN_HASH_KEY"),
		EmailAddress:        os.Getenv("EMAIL"),
		EmailPassword:       os.Getenv("PASSWORD"),
		SessionSecret:       os.Getenv("SESSION_SECRET"),
		GitHubClientID:      os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:  os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		EnableSwagger:       os.Getenv("ENABLE_SWAGGER") == "true",
		EnableMetrics:       os.Getenv("ENABLE_METRICS") == "true",
		RateLimit:           LoadRateLimitConfig(),
	}
}

func (c Config) Validate() error {
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return fmt.Errorf("ENVIRONMENT must be development, test, or production")
	}
	for index, origin := range c.AllowedOrigins {
		parsed, err := url.ParseRequestURI(origin)
		if origin != "*" && (err != nil || parsed.Scheme != "http" && parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil) {
			return fmt.Errorf("ALLOWED_ORIGINS[%d] is invalid: %q", index, origin)
		}
	}
	required := []struct{ name, value string }{
		{"ALLOWED_ORIGINS", first(c.AllowedOrigins)},
		{"BASE_URL", c.BaseURL},
		{"DATABASE_URL", c.DatabaseURL},
		{"JWT_ACCESS_SECRET", c.JWTAccessSecret},
		{"REFRESH_TOKEN_HASH_KEY", c.RefreshTokenHashKey},
		{"EMAIL", c.EmailAddress},
		{"PASSWORD", c.EmailPassword},
		{"SESSION_SECRET", c.SessionSecret},
	}
	for _, setting := range required {
		if setting.value == "" {
			return fmt.Errorf("%s is required", setting.name)
		}
	}
	if (c.GitHubClientID == "") != (c.GitHubClientSecret == "") {
		return fmt.Errorf("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be configured together")
	}
	if (c.GoogleClientID == "") != (c.GoogleClientSecret == "") {
		return fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be configured together")
	}
	if c.Environment == "production" {
		secrets := []struct{ name, value string }{
			{"JWT_ACCESS_SECRET", c.JWTAccessSecret},
			{"REFRESH_TOKEN_HASH_KEY", c.RefreshTokenHashKey},
			{"SESSION_SECRET", c.SessionSecret},
			{"RATE_LIMIT_KEY", c.RateLimit.Key},
		}
		for _, secret := range secrets {
			if len(secret.value) < 32 {
				return fmt.Errorf("%s must be at least 32 bytes in production", secret.name)
			}
		}
	}
	return c.RateLimit.Validate(c.Environment)
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
