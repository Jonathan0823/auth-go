package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestLogger(output *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, nil))
}

func TestMetricsMiddlewareUsesRouteTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics := NewMetrics(nil)
	router := gin.New()
	router.Use(metrics.Middleware())
	router.GET("/users/:id", func(c *gin.Context) {})

	for _, id := range []string{"1", "2"} {
		request := httptest.NewRequest(http.MethodGet, "/users/"+id, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
	}

	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	var requestCount float64
	for _, family := range families {
		if family.GetName() != "auth_go_http_requests_total" {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string)
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if labels["route"] != "/users/:id" {
				t.Fatalf("route label = %q, want /users/:id", labels["route"])
			}
			requestCount += metric.GetCounter().GetValue()
		}
	}
	if requestCount != 2 {
		t.Fatalf("request count = %v, want 2", requestCount)
	}
}

func TestMetricsRouteIsOptIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			metrics := NewMetrics(nil)
			router := gin.New()
			if enabled {
				router.Use(metrics.Middleware())
				router.GET("/ping", func(c *gin.Context) {})
			}
			RegisterMetricsRoute(router, metrics, enabled)
			if enabled {
				pingRequest := httptest.NewRequest(http.MethodGet, "/ping", nil)
				pingResponse := httptest.NewRecorder()
				router.ServeHTTP(pingResponse, pingRequest)
			}
			request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			wantStatus := http.StatusNotFound
			if enabled {
				wantStatus = http.StatusOK
			}
			if response.Code != wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, wantStatus)
			}
			if enabled && !strings.Contains(response.Body.String(), "auth_go_http_requests_total") {
				t.Fatal("metrics response does not contain HTTP request metric")
			}
		})
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
