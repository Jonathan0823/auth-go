package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/platform"
)

func TestMetricsMiddlewareRecordsMatchedAndUnmatchedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics := platform.NewMetrics(nil)
	router := gin.New()
	router.Use(Metrics(metrics))
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	matched := httptest.NewRecorder()
	router.ServeHTTP(matched, httptest.NewRequest(http.MethodGet, "/health", nil))
	if matched.Code != http.StatusNoContent {
		t.Fatalf("matched status = %d", matched.Code)
	}

	unmatched := httptest.NewRecorder()
	router.ServeHTTP(unmatched, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if unmatched.Code != http.StatusNotFound {
		t.Fatalf("unmatched status = %d", unmatched.Code)
	}

	families, err := metrics.Registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	if len(families) == 0 {
		t.Fatal("metrics registry is empty")
	}
}

func TestMetricsMiddlewareWithoutMetricsIsTransparent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Metrics(nil))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}
