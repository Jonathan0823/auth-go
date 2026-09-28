package middleware

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/gin-gonic/gin"
)

func TestErrorHandlerRateLimitResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name       string
		err        error
		status     int
		retryAfter string
	}{
		{"limited", &service.RateLimitError{RetryAfter: 2500 * time.Millisecond}, http.StatusTooManyRequests, "3"},
		{"backend unavailable", &service.RateLimitError{Unavailable: true}, http.StatusServiceUnavailable, "30"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
			router.GET("/", func(c *gin.Context) { _ = c.Error(test.err) })
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != test.status || response.Header().Get("Retry-After") != test.retryAfter || response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status=%d headers=%v", response.Code, response.Header())
			}
		})
	}
}

func TestMapError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"invalid input", fmt.Errorf("bad field: %w", domain.ErrInvalidInput), http.StatusBadRequest, "invalid input"},
		{"unauthenticated", domain.ErrUnauthenticated, http.StatusUnauthorized, "unauthorized"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"not found", domain.ErrNotFound, http.StatusNotFound, "not found"},
		{"conflict", domain.ErrConflict, http.StatusConflict, "conflict"},
		{"rate limited", &service.RateLimitError{RetryAfter: 10 * time.Second}, http.StatusTooManyRequests, "too many requests"},
		{"rate limit backend unavailable", &service.RateLimitError{Unavailable: true}, http.StatusServiceUnavailable, "authentication temporarily unavailable"},
		{"unknown", errors.New("database password leaked"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message := mapError(tt.err)
			if status != tt.status || message != tt.message {
				t.Fatalf("mapError() = (%d, %q), want (%d, %q)", status, message, tt.status, tt.message)
			}
		})
	}
}
