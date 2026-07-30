package platform

import (
	"log/slog"
	"os"
	"strings"
)

const redactedValue = "[REDACTED]"

func NewLogger(level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: ParseLogLevel(level),
	}))
}

func ParseLogLevel(value string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Redact(key string, value any) slog.Attr {
	if isSensitiveKey(key) {
		return slog.String(key, redactedValue)
	}
	return slog.Any(key, value)
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
	for _, sensitive := range []string{"password", "token", "authorization", "cookie", "secret"} {
		if strings.Contains(key, sensitive) {
			return true
		}
	}
	return false
}
