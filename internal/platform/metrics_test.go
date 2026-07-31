package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(output *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}

func TestAuditLoggerRecordsRateLimitMetric(t *testing.T) {
	metrics := NewMetrics(nil)
	logger := NewAuditLogger(slog.New(slog.NewJSONHandler(io.Discard, nil)), metrics)
	logger.Log(context.Background(), AuditEvent{
		Name:    EventAuthRateLimited,
		Outcome: OutcomeDetected,
		Route:   "/api/auth/login",
		Reason:  ReasonRateLimited,
		Policy:  "login_ip",
		Backend: "memory",
	})

	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() == "auth_go_rate_limit_events_total" && len(family.GetMetric()) == 1 {
			return
		}
	}
	t.Fatal("rate-limit metric was not recorded")
}

func TestMetricsNormalizesUnknownValues(t *testing.T) {
	metrics := NewMetrics(nil)
	metrics.RecordSecurityEvent("unknown", "unknown", "unknown", "unknown")
	metrics.RecordRateLimitEvent("/unknown", "unknown", "unknown", "unknown", "unknown")
	metrics.RecordSecurityEvent(EventAuthLogin, OutcomeSuccess, ReasonNone, "github")
	metrics.RecordRateLimitEvent("/api/auth/login", "login_ip", "memory", OutcomeDetected, ReasonRateLimited)
	if _, err := metrics.Registry.Gather(); err != nil {
		t.Fatal(err)
	}
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
		Policy:    "login_ip",
		Backend:   "memory",
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
		"policy":     "login_ip",
		"backend":    "memory",
	} {
		if entry[key] != want {
			t.Fatalf("%s = %v, want %v", key, entry[key], want)
		}
	}
	if strings.Contains(output.String(), "token") || strings.Contains(output.String(), "secret") {
		t.Fatalf("audit log contains sensitive terms: %s", output.String())
	}
}
