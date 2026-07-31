package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/service"
	"github.com/Jonathan0823/auth-go/internal/platform"
	"github.com/gin-gonic/gin"
)

func TestAuditIncludesRequestIDWithoutSensitiveInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var output bytes.Buffer
	handler := &Handler{
		Audit: platform.NewAuditLogger(slog.New(slog.NewJSONHandler(&output, nil)), platform.NewMetrics(nil)),
	}
	router := gin.New()
	router.Use(inhttp.RequestID())
	router.GET("/audit", func(c *gin.Context) {
		handler.auditSuccess(c, platform.EventAuthLogin, "", 0)
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/audit", nil)
	request.Header.Set("X-Request-ID", "11111111-1111-1111-1111-111111111111")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("audit output is not valid JSON: %v", err)
	}
	if entry["request_id"] != request.Header.Get("X-Request-ID") {
		t.Fatalf("request_id = %v, want %s", entry["request_id"], request.Header.Get("X-Request-ID"))
	}
	if bytes.Contains(output.Bytes(), []byte("password")) || bytes.Contains(output.Bytes(), []byte("token")) {
		t.Fatalf("audit output contains sensitive input: %s", output.Bytes())
	}
}

func TestAuditReasonClassifiesErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "replay", err: fmtReplayError(), want: platform.ReasonReplayDetected},
		{name: "validation", err: domain.ErrInvalidInput, want: platform.ReasonValidation},
		{name: "unauthenticated", err: domain.ErrUnauthenticated, want: platform.ReasonUnauthenticated},
		{name: "unknown", err: errors.New("internal details"), want: platform.ReasonInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := auditReason(tt.err); got != tt.want {
				t.Fatalf("auditReason() = %q, want %q", got, tt.want)
			}
		})
	}
}

func fmtReplayError() error {
	return errors.Join(service.ErrRefreshTokenReused, domain.ErrUnauthenticated)
}
