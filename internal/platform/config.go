package platform

import "os"

type Config struct {
	Port           string
	AllowedOrigins string
	Environment    string
	BaseURL        string
	LogLevel       string
}

func LoadConfig() Config {
	return Config{
		Port:           os.Getenv("PORT"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		Environment:    os.Getenv("ENVIRONMENT"),
		BaseURL:        os.Getenv("BASE_URL"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
	}
}
