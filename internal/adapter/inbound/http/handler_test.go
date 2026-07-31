package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	middleware "github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/middleware"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

type handlerAuthService struct {
	registerErr     error
	loginErr        error
	forgotErr       error
	verifyCreateErr error
	verifyErr       error
	resetErr        error
	refreshErr      error
	logoutErr       error
	registerCalls   int
	loginCalls      int
	forgotCalls     int
}

func (f *handlerAuthService) Register(context.Context, domain.User) error {
	f.registerCalls++
	return f.registerErr
}
func (f *handlerAuthService) Login(context.Context, domain.User) (string, string, error) {
	f.loginCalls++
	return "access", "refresh", f.loginErr
}
func (f *handlerAuthService) ForgotPassword(context.Context, string) error {
	f.forgotCalls++
	return f.forgotErr
}
func (f *handlerAuthService) CreateVerifyEmail(context.Context, string) error {
	return f.verifyCreateErr
}
func (f *handlerAuthService) VerifyEmail(context.Context, string) error { return f.verifyErr }
func (f *handlerAuthService) ResetPassword(context.Context, string, string) error {
	return f.resetErr
}
func (f *handlerAuthService) RefreshTokens(context.Context, string, string, string) (string, string, error) {
	return "access", "refresh", f.refreshErr
}
func (f *handlerAuthService) Logout(context.Context, string) error { return f.logoutErr }

type handlerUserService struct {
	user        *domain.User
	users       []*domain.User
	getByIDErr  error
	getEmailErr error
	getAllErr   error
	updateErr   error
	deleteErr   error
}

func (f *handlerUserService) GetByID(context.Context, int) (*domain.User, error) {
	return f.user, f.getByIDErr
}
func (f *handlerUserService) GetByEmail(context.Context, string) (*domain.User, error) {
	return f.user, f.getEmailErr
}
func (f *handlerUserService) GetAll(context.Context) ([]*domain.User, error) {
	return f.users, f.getAllErr
}
func (f *handlerUserService) Update(context.Context, int, domain.UpdateUserCommand) error {
	return f.updateErr
}
func (f *handlerUserService) Delete(context.Context, int, int) error { return f.deleteErr }

type handlerOAuthService struct {
	user         *domain.User
	err          error
	beginCall    bool
	callbackCall bool
}

func (f *handlerOAuthService) BeginAuth(http.ResponseWriter, *http.Request, string) {
	f.beginCall = true
}
func (f *handlerOAuthService) OAuthLogin(context.Context, http.ResponseWriter, *http.Request, string) (*domain.User, error) {
	f.callbackCall = true
	return f.user, f.err
}

type handlerTokenService struct{}

func (handlerTokenService) GenerateAccessToken(domain.User) (string, string, error) {
	return "access", "jti", nil
}
func (handlerTokenService) ValidateAccessToken(string) (map[string]interface{}, error) {
	return nil, nil
}
func (handlerTokenService) GenerateRefreshToken() (string, []byte, error) {
	return "refresh", []byte("hash"), nil
}
func (handlerTokenService) HashRefreshToken(string) ([]byte, error) { return []byte("hash"), nil }

func newHandlerTest(auth port.AuthService, user port.UserService, oauth port.OAuthService) *Handler {
	return &Handler{Svc: port.Service{Auth: auth, User: user, OAuth: oauth}, Tokens: handlerTokenService{}}
}

func newErrorRouter(h gin.HandlerFunc) *gin.Engine {
	return newPathRouter("/", h)
}

func newPathRouter(path string, h gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
	router.Any(path, h)
	return router
}

func request(t *testing.T, router http.Handler, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var data io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		data = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, data)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func validCredentials() map[string]string {
	return map[string]string{"email": "user@example.com", "password": "password-123"}
}

func TestAuthHandlersSuccess(t *testing.T) {
	auth := &handlerAuthService{}
	h := newHandlerTest(auth, &handlerUserService{}, &handlerOAuthService{})

	if response := request(t, newErrorRouter(h.Register), http.MethodPost, "/", validCredentials(), nil); response.Code != http.StatusOK {
		t.Fatalf("register status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.Login), http.MethodPost, "/", validCredentials(), nil); response.Code != http.StatusOK || len(response.Result().Cookies()) != 2 {
		t.Fatalf("login response = %d, cookies=%v", response.Code, response.Result().Cookies())
	}
	if response := request(t, newErrorRouter(h.Logout), http.MethodPost, "/", nil, &http.Cookie{Name: "refresh_token", Value: "refresh"}); response.Code != http.StatusOK {
		t.Fatalf("logout status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.Refresh), http.MethodPost, "/", nil, &http.Cookie{Name: "refresh_token", Value: "refresh"}); response.Code != http.StatusOK {
		t.Fatalf("refresh status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.VerifyEmail), http.MethodGet, "/?id="+uuid.NewString(), nil, nil); response.Code != http.StatusOK {
		t.Fatalf("verify status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ResendVerifyEmail), http.MethodPost, "/?email=user@example.com", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("resend status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ForgotPassword), http.MethodPost, "/", map[string]string{"email": "user@example.com"}, nil); response.Code != http.StatusOK {
		t.Fatalf("forgot status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ResetPassword), http.MethodPost, "/", map[string]string{"id": uuid.NewString(), "password": "new-password-123"}, nil); response.Code != http.StatusOK {
		t.Fatalf("reset status = %d", response.Code)
	}
	if auth.registerCalls != 1 || auth.loginCalls != 1 || auth.forgotCalls != 1 {
		t.Fatalf("auth calls = register:%d login:%d forgot:%d", auth.registerCalls, auth.loginCalls, auth.forgotCalls)
	}
}

func TestAuthHandlersValidationAndErrors(t *testing.T) {
	auth := &handlerAuthService{}
	h := newHandlerTest(auth, &handlerUserService{}, &handlerOAuthService{})
	if response := request(t, newErrorRouter(h.Register), http.MethodPost, "/", map[string]string{}, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid register status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.Login), http.MethodPost, "/", map[string]string{}, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid login status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ResendVerifyEmail), http.MethodPost, "/", nil, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid resend status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ForgotPassword), http.MethodPost, "/", map[string]string{}, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid forgot status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.ResetPassword), http.MethodPost, "/", map[string]string{}, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid reset status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.Logout), http.MethodPost, "/", nil, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("missing logout status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.Refresh), http.MethodPost, "/", nil, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("missing refresh status = %d", response.Code)
	}

	auth.registerErr = domain.ErrConflict
	if response := request(t, newErrorRouter(h.Register), http.MethodPost, "/", validCredentials(), nil); response.Code != http.StatusConflict {
		t.Fatalf("register error status = %d", response.Code)
	}
	auth.registerErr = nil
	auth.loginErr = domain.ErrUnauthenticated
	if response := request(t, newErrorRouter(h.Login), http.MethodPost, "/", validCredentials(), nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("login error status = %d", response.Code)
	}
	auth.loginErr = nil
	auth.forgotErr = errors.New("forgot failure")
	if response := request(t, newErrorRouter(h.ForgotPassword), http.MethodPost, "/", map[string]string{"email": "user@example.com"}, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("forgot error status = %d", response.Code)
	}
}

func TestOAuthHandlers(t *testing.T) {
	oauth := &handlerOAuthService{user: &domain.User{ID: 1, Email: "oauth@example.com"}}
	h := newHandlerTest(&handlerAuthService{}, &handlerUserService{}, oauth)
	if response := request(t, newErrorRouter(h.OAuthLogin), http.MethodGet, "/", nil, nil); response.Code != http.StatusOK || !oauth.beginCall {
		t.Fatalf("oauth begin response = %d, called=%v", response.Code, oauth.beginCall)
	}
	if response := request(t, newErrorRouter(h.OAuthCallback), http.MethodGet, "/", nil, nil); response.Code != http.StatusOK || !oauth.callbackCall {
		t.Fatalf("oauth callback response = %d, called=%v", response.Code, oauth.callbackCall)
	}
	oauth.err = errors.New("provider failure")
	if response := request(t, newErrorRouter(h.OAuthCallback), http.MethodGet, "/", nil, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("oauth error status = %d", response.Code)
	}
}

func TestUserHandlers(t *testing.T) {
	user := &domain.User{ID: 1, Email: "user@example.com", Username: "user"}
	users := &handlerUserService{user: user, users: []*domain.User{user}}
	h := newHandlerTest(&handlerAuthService{}, users, &handlerOAuthService{})
	withUser := func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("user", map[string]interface{}{"id": float64(1), "username": "user", "email": "user@example.com"})
			handler(c)
		}
	}
	if response := request(t, newPathRouter("/:id", h.GetUserByID), http.MethodGet, "/1", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("get by ID status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.GetAllUsers), http.MethodGet, "/", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("get all status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.GetUserByEmail), http.MethodGet, "/?email=user@example.com", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("get by email status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(withUser(h.UpdateUser)), http.MethodPatch, "/", map[string]any{"id": 1, "email": "user@example.com"}, nil); response.Code != http.StatusOK {
		t.Fatalf("update status = %d", response.Code)
	}
	if response := request(t, newPathRouter("/:id", withUser(h.DeleteUser)), http.MethodDelete, "/1", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("delete status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(withUser(h.GetCurrentUser)), http.MethodGet, "/", nil, nil); response.Code != http.StatusOK {
		t.Fatalf("current user status = %d", response.Code)
	}
	if response := request(t, newPathRouter("/:id", h.GetUserByID), http.MethodGet, "/0", nil, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid ID status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.GetUserByEmail), http.MethodGet, "/", nil, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("missing email status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.UpdateUser), http.MethodPatch, "/", map[string]any{"id": 1, "email": "bad"}, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid update status = %d", response.Code)
	}
	if response := request(t, newPathRouter("/:id", h.DeleteUser), http.MethodDelete, "/0", nil, nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid delete status = %d", response.Code)
	}
	if response := request(t, newErrorRouter(h.GetCurrentUser), http.MethodGet, "/", nil, nil); response.Code != http.StatusUnauthorized {
		t.Fatalf("missing current user status = %d", response.Code)
	}

	users.getByIDErr = errors.New("lookup failure")
	if response := request(t, newPathRouter("/:id", h.GetUserByID), http.MethodGet, "/1", nil, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("get by ID error status = %d", response.Code)
	}
	users.getAllErr = errors.New("list failure")
	if response := request(t, newErrorRouter(h.GetAllUsers), http.MethodGet, "/", nil, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("get all error status = %d", response.Code)
	}
	users.getEmailErr = errors.New("email failure")
	if response := request(t, newErrorRouter(h.GetUserByEmail), http.MethodGet, "/?email=user@example.com", nil, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("get by email error status = %d", response.Code)
	}
	users.updateErr = errors.New("update failure")
	if response := request(t, newErrorRouter(withUser(h.UpdateUser)), http.MethodPatch, "/", map[string]any{"id": 1, "email": "user@example.com"}, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("update error status = %d", response.Code)
	}
	users.deleteErr = errors.New("delete failure")
	if response := request(t, newPathRouter("/:id", withUser(h.DeleteUser)), http.MethodDelete, "/1", nil, nil); response.Code != http.StatusInternalServerError {
		t.Fatalf("delete error status = %d", response.Code)
	}
}

var _ port.AuthService = (*handlerAuthService)(nil)
var _ port.UserService = (*handlerUserService)(nil)
var _ port.OAuthService = (*handlerOAuthService)(nil)
var _ port.TokenService = (handlerTokenService{})
