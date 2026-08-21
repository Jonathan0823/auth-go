//go:build integration

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	inhttp "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http"
	outjwt "github.com/Jonathan0823/auth-go/internal/adapter/outbound/jwt"
	outpassword "github.com/Jonathan0823/auth-go/internal/adapter/outbound/password"
	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/core/service"
)

type integrationEmailSender struct{}

func (integrationEmailSender) Send(string, string, string) error { return nil }

func setupAuthServer(t *testing.T) (*gin.Engine, *pgxpool.Pool, port.Repository, port.TokenService) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("ping database: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := outpostgres.NewRepository(pool)
	tokens := outjwt.NewTokenService(os.Getenv("JWT_ACCESS_SECRET"), os.Getenv("REFRESH_TOKEN_HASH_KEY"))
	svc := service.New(repo, tokens, integrationEmailSender{}, outpassword.NewHasher(), "http://localhost:8080")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	inhttp.RegisterRoutes(router, inhttp.NewHandler(svc, tokens, nil, inhttp.HandlerConfig{}), logger)
	return router, pool, repo, tokens
}

func TestAuthHTTPFlow(t *testing.T) {
	router, pool, repo, tokens := setupAuthServer(t)
	email := fmt.Sprintf("auth-%s@example.com", uuid.NewString())
	password := "correct-password-123"
	cleanupUser(t, pool, email)

	register := doJSON(t, router, http.MethodPost, "/api/auth/register", map[string]string{
		"email":    email,
		"password": password,
	})
	if register.StatusCode != http.StatusOK {
		t.Fatalf("register status = %d, want %d", register.StatusCode, http.StatusOK)
	}
	register.Body.Close()

	var passwordHash string
	if err := pool.QueryRow(context.Background(), `SELECT password FROM users WHERE email = $1`, email).Scan(&passwordHash); err != nil {
		t.Fatalf("read password hash: %v", err)
	}
	if !strings.HasPrefix(passwordHash, "$argon2id$") {
		t.Fatalf("password hash = %q, want Argon2id format", passwordHash)
	}

	login := doJSON(t, router, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if login.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", login.StatusCode, http.StatusOK)
	}
	access := responseCookie(t, login, "access_token")
	refresh := responseCookie(t, login, "refresh_token")
	assertSecureCookie(t, access)
	assertSecureCookie(t, refresh)
	if strings.Contains(refresh.Value, ".") {
		t.Fatal("refresh token must not contain JWT separators")
	}
	login.Body.Close()

	protected := doJSON(t, router, http.MethodGet, "/api/user/me", nil, access)
	if protected.StatusCode != http.StatusOK {
		t.Fatalf("protected endpoint status = %d, want %d", protected.StatusCode, http.StatusOK)
	}
	protected.Body.Close()

	oldHash, err := tokens.HashRefreshToken(refresh.Value)
	if err != nil {
		t.Fatalf("hash original refresh token: %v", err)
	}
	oldRecord, err := repo.Auth().GetRefreshTokenByHash(context.Background(), oldHash)
	if err != nil || oldRecord == nil {
		t.Fatalf("find original refresh token: record=%v err=%v", oldRecord, err)
	}

	rotated := doJSON(t, router, http.MethodPost, "/api/auth/refresh", nil, refresh)
	if rotated.StatusCode != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d", rotated.StatusCode, http.StatusOK)
	}
	newRefresh := responseCookie(t, rotated, "refresh_token")
	if newRefresh.Value == refresh.Value {
		t.Fatal("refresh token was not rotated")
	}
	rotated.Body.Close()

	usedRecord, err := repo.Auth().GetRefreshTokenByHash(context.Background(), oldHash)
	if err != nil || usedRecord == nil || usedRecord.UsedAt == nil {
		t.Fatalf("original token was not marked used: record=%v err=%v", usedRecord, err)
	}
	newHash, err := tokens.HashRefreshToken(newRefresh.Value)
	if err != nil {
		t.Fatalf("hash new refresh token: %v", err)
	}
	newRecord, err := repo.Auth().GetRefreshTokenByHash(context.Background(), newHash)
	if err != nil || newRecord == nil {
		t.Fatalf("find rotated refresh token: record=%v err=%v", newRecord, err)
	}
	if bytes.Equal(newRecord.TokenHash, []byte(newRefresh.Value)) {
		t.Fatal("raw refresh token was persisted")
	}

	replay := doJSON(t, router, http.MethodPost, "/api/auth/refresh", nil, refresh)
	if replay.StatusCode != http.StatusUnauthorized {
		t.Fatalf("replay status = %d, want %d", replay.StatusCode, http.StatusUnauthorized)
	}
	replay.Body.Close()

	revokedRecord, err := repo.Auth().GetRefreshTokenByHash(context.Background(), newHash)
	if err != nil || revokedRecord == nil || revokedRecord.RevokedAt == nil {
		t.Fatalf("token family was not revoked after replay: record=%v err=%v", revokedRecord, err)
	}

	loginAgain := doJSON(t, router, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if loginAgain.StatusCode != http.StatusOK {
		t.Fatalf("second login status = %d, want %d", loginAgain.StatusCode, http.StatusOK)
	}
	refreshAgain := responseCookie(t, loginAgain, "refresh_token")
	loginAgain.Body.Close()

	logout := doJSON(t, router, http.MethodPost, "/api/auth/logout", nil, refreshAgain)
	if logout.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", logout.StatusCode, http.StatusOK)
	}
	for _, cookie := range logout.Cookies() {
		if (cookie.Name == "access_token" || cookie.Name == "refresh_token") && cookie.MaxAge != -1 {
			t.Fatalf("cookie %q MaxAge = %d, want -1", cookie.Name, cookie.MaxAge)
		}
	}
	logout.Body.Close()

	afterLogout := doJSON(t, router, http.MethodPost, "/api/auth/refresh", nil, refreshAgain)
	if afterLogout.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status = %d, want %d", afterLogout.StatusCode, http.StatusUnauthorized)
	}
	afterLogout.Body.Close()
}

func TestAuthRefreshConcurrentUse(t *testing.T) {
	router, pool, _, _ := setupAuthServer(t)
	email := fmt.Sprintf("concurrent-%s@example.com", uuid.NewString())
	password := "correct-password-123"
	cleanupUser(t, pool, email)

	register := doJSON(t, router, http.MethodPost, "/api/auth/register", map[string]string{
		"email":    email,
		"password": password,
	})
	if register.StatusCode != http.StatusOK {
		t.Fatalf("register status = %d, want %d", register.StatusCode, http.StatusOK)
	}
	register.Body.Close()

	login := doJSON(t, router, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    email,
		"password": password,
	})
	if login.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d, want %d", login.StatusCode, http.StatusOK)
	}
	refresh := responseCookie(t, login, "refresh_token")
	login.Body.Close()

	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := doJSON(t, router, http.MethodPost, "/api/auth/refresh", nil, refresh)
			statuses <- response.StatusCode
			response.Body.Close()
		}()
	}
	wg.Wait()
	close(statuses)

	var success, unauthorized int
	for status := range statuses {
		switch status {
		case http.StatusOK:
			success++
		case http.StatusUnauthorized:
			unauthorized++
		}
	}
	if success != 1 || unauthorized != 1 {
		t.Fatalf("concurrent refresh results = success:%d unauthorized:%d, want 1:1", success, unauthorized)
	}
}

func cleanupUser(t *testing.T, pool *pgxpool.Pool, email string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE email = $1`, email); err != nil {
			t.Logf("cleanup user %q: %v", email, err)
		}
	})
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any, cookies ...*http.Cookie) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Result()
}

func responseCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not include %q cookie", name)
	return nil
}

func assertSecureCookie(t *testing.T, cookie *http.Cookie) {
	t.Helper()
	if !cookie.HttpOnly {
		t.Fatalf("cookie %q is not HttpOnly", cookie.Name)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie %q SameSite = %v, want Lax", cookie.Name, cookie.SameSite)
	}
}
