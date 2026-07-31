package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeDBPinger struct {
	err error
}

func (f fakeDBPinger) Ping(context.Context) error {
	return f.err
}

func TestHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		path       string
		pinger     DBPinger
		wantStatus int
		wantBody   string
	}{
		{name: "live without database", path: "/health/live", pinger: fakeDBPinger{err: errors.New("database unavailable")}, wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},
		{name: "ready with database", path: "/health/ready", pinger: fakeDBPinger{}, wantStatus: http.StatusOK, wantBody: `{"status":"ready"}`},
		{name: "ready without database", path: "/health/ready", pinger: fakeDBPinger{err: errors.New("database unavailable")}, wantStatus: http.StatusServiceUnavailable, wantBody: `{"status":"not_ready"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			RegisterHealthRoutes(router, tt.pinger)
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Body.String(); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestHealthReadinessHandlesNilDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterHealthRoutes(router, nil)
	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
