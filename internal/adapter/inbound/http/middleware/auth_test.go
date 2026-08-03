package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

type middlewareTokenService struct {
	claims map[string]any
	err    error
}

func (t middlewareTokenService) GenerateAccessToken(domain.User) (string, string, error) {
	return "", "", nil
}
func (t middlewareTokenService) ValidateAccessToken(string) (map[string]any, error) {
	return t.claims, t.err
}
func (t middlewareTokenService) GenerateRefreshToken() (string, []byte, error) {
	return "", nil, nil
}
func (t middlewareTokenService) HashRefreshToken(string) ([]byte, error) { return nil, nil }

func TestAuthMiddlewareRejectsMissingAndInvalidTokens(t *testing.T) {
	for name, tokens := range map[string]middlewareTokenService{
		"missing": {},
		"invalid": {err: domain.ErrUnauthenticated},
	} {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
			router.Use(NewAuthMiddleware(tokens).Handler())
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewAuthMiddleware(middlewareTokenService{
		claims: map[string]any{"id": float64(1)},
	}).Handler())
	router.GET("/", func(c *gin.Context) {
		if _, exists := c.Get("user"); !exists {
			t.Error("user claims were not stored")
		}
		c.Status(http.StatusNoContent)
	})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "access_token", Value: "valid"})
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}
