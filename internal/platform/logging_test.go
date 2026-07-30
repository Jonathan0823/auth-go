package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := map[string]slog.Level{
		"DEBUG":     slog.LevelDebug,
		"info":      slog.LevelInfo,
		" WARNING ": slog.LevelWarn,
		"ERROR":     slog.LevelError,
		"invalid":   slog.LevelInfo,
		"":          slog.LevelInfo,
	}
	for input, want := range tests {
		if got := ParseLogLevel(input); got != want {
			t.Errorf("ParseLogLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestRedactSensitiveAttributes(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	logger.LogAttrs(context.Background(), slog.LevelInfo, "request",
		Redact("password", "super-secret"),
		Redact("Access-Token", "access-secret"),
		Redact("user_agent", "test-client"),
	)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
	if entry["password"] != redactedValue || entry["Access-Token"] != redactedValue {
		t.Fatalf("sensitive attributes were not redacted: %v", entry)
	}
	if entry["user_agent"] != "test-client" {
		t.Fatalf("non-sensitive attribute was changed: %v", entry["user_agent"])
	}
	if bytes.Contains(output.Bytes(), []byte("super-secret")) || bytes.Contains(output.Bytes(), []byte("access-secret")) {
		t.Fatalf("sensitive value appeared in log output: %s", output.Bytes())
	}
}
