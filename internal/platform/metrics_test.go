package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(output *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}

func TestAuditLoggerUsesSafeFields(t *testing.T) {
	var output bytes.Buffer
	logger := NewAuditLogger(
		newTestLogger(&output),
		NewMetrics(nil),
	)
	logger.Log(context.Background(), AuditEvent{
		Name:      EventAuthRefreshReplay,
		Outcome:   OutcomeDetected,
		RequestID: "request-id",
		UserID:    42,
		Provider:  "github",
		Reason:    ReasonReplayDetected,
	})

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("log output is not valid JSON: %v", err)
	}
	for key, want := range map[string]any{
		"event":      EventAuthRefreshReplay,
		"outcome":    OutcomeDetected,
		"request_id": "request-id",
		"user_id":    float64(42),
		"provider":   "github",
		"reason":     ReasonReplayDetected,
	} {
		if entry[key] != want {
			t.Fatalf("%s = %v, want %v", key, entry[key], want)
		}
	}
	if strings.Contains(output.String(), "token") || strings.Contains(output.String(), "secret") {
		t.Fatalf("audit log contains sensitive terms: %s", output.String())
	}
}
