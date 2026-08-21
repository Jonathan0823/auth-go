package platform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS(Config{AllowedOrigins: []string{"https://app.example.com"}}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://app.example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestRunServerShutsDownWhenContextIsCancelled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	if err := RunServer(ctx, gin.New(), Config{Port: "0"}); err != nil {
		t.Fatalf("RunServer() error = %v", err)
	}
}

func TestHTTPServerHasProductionTimeouts(t *testing.T) {
	server := newHTTPServer(http.NotFoundHandler(), "9090")
	if server.Addr != ":9090" {
		t.Fatalf("Addr = %q", server.Addr)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 15*time.Second || server.WriteTimeout != 30*time.Second || server.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected timeouts: %+v", server)
	}
}
