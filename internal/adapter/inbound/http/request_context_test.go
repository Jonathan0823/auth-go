package http

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetUserRejectsMalformedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, claims := range []any{
		map[string]any{"id": "1", "username": "user", "email": "user@example.com"},
		map[string]any{"id": float64(1), "username": 1, "email": "user@example.com"},
		"invalid claims",
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user", claims)
		if _, err := GetUser(c); err == nil {
			t.Errorf("GetUser(%#v) succeeded, want malformed-claims error", claims)
		}
	}
}
