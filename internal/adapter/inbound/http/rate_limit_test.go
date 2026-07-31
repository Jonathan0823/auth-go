package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

type fakeRateLimitStore struct {
	decision port.RateLimitDecision
	err      error
}

func (s fakeRateLimitStore) Allow(context.Context, string, port.RateLimitPolicy) (port.RateLimitDecision, error) {
	return s.decision, s.err
}
func (fakeRateLimitStore) Reset(context.Context, string) error { return nil }
func (fakeRateLimitStore) Backend() string                     { return "memory" }
func (fakeRateLimitStore) Close() error                        { return nil }

func TestRateLimitResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{RateLimiter: platform.NewRateLimiter(
		fakeRateLimitStore{decision: port.RateLimitDecision{RetryAfter: 10 * time.Second}},
		"test-key",
		map[string]port.RateLimitPolicy{"login_ip": {Name: "login_ip", Limit: 1, Window: time.Minute}},
	)}
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		if handler.allowRateLimit(c, ipRateLimit("login_ip", c)) {
			c.Status(http.StatusNoContent)
		}
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", response.Code)
	}
	if response.Header().Get("Retry-After") != "10" {
		t.Fatalf("Retry-After = %q, want 10", response.Header().Get("Retry-After"))
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", response.Header().Get("Cache-Control"))
	}
	if response.Body.String() != `{"error":"too many requests"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestRateLimitUsesDirectPeerWithoutTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		t.Fatal(err)
	}
	var clientIP string
	router.GET("/", func(c *gin.Context) {
		clientIP = c.ClientIP()
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.20")
	router.ServeHTTP(httptest.NewRecorder(), request)
	if clientIP != "192.0.2.10" {
		t.Fatalf("client IP = %q, want direct peer", clientIP)
	}
}

func TestRateLimitUsesForwardedIPFromTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := router.SetTrustedProxies([]string{"192.0.2.0/24"}); err != nil {
		t.Fatal(err)
	}
	var clientIP string
	router.GET("/", func(c *gin.Context) {
		clientIP = c.ClientIP()
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "192.0.2.10:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.20")
	router.ServeHTTP(httptest.NewRecorder(), request)
	if clientIP != "198.51.100.20" {
		t.Fatalf("client IP = %q, want forwarded client", clientIP)
	}
}

func TestRateLimitBackendFailureDoesNotLeakError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{RateLimiter: platform.NewRateLimiter(
		fakeRateLimitStore{err: errors.New("redis password leaked")},
		"test-key",
		map[string]port.RateLimitPolicy{"login_ip": {Name: "login_ip", Limit: 1, Window: time.Minute}},
	)}
	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		if handler.allowRateLimit(c, ipRateLimit("login_ip", c)) {
			c.Status(http.StatusNoContent)
		}
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.Code)
	}
	if response.Body.String() != `{"error":"authentication temporarily unavailable"}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}
