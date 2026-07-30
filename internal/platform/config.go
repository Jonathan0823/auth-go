package platform

import "os"

type Config struct {
	Port           string
	AllowedOrigins string
	Environment    string
}

func LoadConfig() Config {
	return Config{
		Port:           os.Getenv("PORT"),
		AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		Environment:    os.Getenv("ENVIRONMENT"),
	}
}
