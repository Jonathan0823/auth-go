package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		requestID  string
		wantSameID bool
	}{
		{name: "generates missing id"},
		{name: "preserves valid id", requestID: "550e8400-e29b-41d4-a716-446655440000", wantSameID: true},
		{name: "replaces invalid id", requestID: "not-a-uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			var contextID string
			r.Use(RequestID())
			r.GET("/", func(c *gin.Context) {
				contextID = RequestIDFromContext(c.Request.Context())
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.requestID != "" {
				req.Header.Set(requestIDHeader, tt.requestID)
			}
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)

			responseID := res.Header().Get(requestIDHeader)
			if responseID != contextID || responseID == "" {
				t.Fatalf("request ID response=%q context=%q", responseID, contextID)
			}
			if _, err := uuid.Parse(responseID); err != nil {
				t.Fatalf("request ID is not a UUID: %q", responseID)
			}
			if tt.wantSameID && responseID != tt.requestID {
				t.Fatalf("request ID = %q, want %q", responseID, tt.requestID)
			}
			if !tt.wantSameID && responseID == tt.requestID {
				t.Fatalf("invalid request ID was preserved: %q", responseID)
			}
		})
	}
}

func TestRequestLoggerWritesSafeJSONFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	r := gin.New()
	r.Use(RequestID(), RequestLogger(logger))
	r.POST("/users", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"password":"do-not-log"}`))
	req.Header.Set(requestIDHeader, "550e8400-e29b-41d4-a716-446655440000")
	req.Header.Set("User-Agent", "test-client")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("request log is not valid JSON: %v", err)
	}
	if entry["msg"] != "http request" || entry["request_id"] != req.Header.Get(requestIDHeader) {
		t.Fatalf("unexpected request log: %v", entry)
	}
	if entry["method"] != http.MethodPost || entry["path"] != "/users" || entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("missing request fields: %v", entry)
	}
	if _, ok := entry["duration"]; !ok {
		t.Fatalf("request duration is missing: %v", entry)
	}
	if _, ok := entry["body"]; ok || bytes.Contains(output.Bytes(), []byte("do-not-log")) {
		t.Fatalf("request body was logged: %s", output.Bytes())
	}
}

func TestErrorHandlerLogsOnlyServerErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		err           error
		wantStatus    int
		wantErrorLogs int
	}{
		{name: "expected client error", err: domain.ErrInvalidInput, wantStatus: http.StatusBadRequest},
		{name: "application error", err: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantErrorLogs: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
			r := gin.New()
			r.Use(RequestID(), RequestLogger(logger))
			api := r.Group("/api")
			api.Use(ErrorHandler(logger))
			api.GET("/test", func(c *gin.Context) {
				c.Error(tt.err)
			})

			res := httptest.NewRecorder()
			r.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/test", nil))
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}

			errorLogs := 0
			for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
				var entry map[string]any
				if err := json.Unmarshal([]byte(line), &entry); err != nil {
					t.Fatalf("log is not valid JSON: %v", err)
				}
				if entry["level"] == "ERROR" {
					errorLogs++
					if entry["request_id"] == "" || entry["error"] != "internal server error" {
						t.Fatalf("missing error context: %v", entry)
					}
				}
			}
			if errorLogs != tt.wantErrorLogs {
				t.Fatalf("error log count = %d, want %d; output=%s", errorLogs, tt.wantErrorLogs, output.String())
			}
		})
	}
}
