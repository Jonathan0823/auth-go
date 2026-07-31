package platform

import "os"

type Config struct {
	Port               string
	AllowedOrigins     string
	Environment        string
	BaseURL            string
	LogLevel           string
	SessionSecret      string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	EnableSwagger      bool
}

func LoadConfig() Config {
	return Config{
		Port:               os.Getenv("PORT"),
		AllowedOrigins:     os.Getenv("ALLOWED_ORIGINS"),
		Environment:        os.Getenv("ENVIRONMENT"),
		BaseURL:            os.Getenv("BASE_URL"),
		LogLevel:           os.Getenv("LOG_LEVEL"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		EnableSwagger:      os.Getenv("ENABLE_SWAGGER") == "true",
	}
}
