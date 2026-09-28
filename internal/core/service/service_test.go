package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
	"github.com/Jonathan0823/auth-go/internal/core/ratelimit"
)

var errFake = errors.New("fake failure")

type fakeUserRepository struct {
	getByIDFn        func(context.Context, int) (*domain.User, error)
	getByEmailFn     func(context.Context, string, bool) (*domain.User, error)
	createFn         func(context.Context, domain.User) error
	getAllFn         func(context.Context) ([]*domain.User, error)
	updateFn         func(context.Context, domain.UpdateUserCommand) error
	deleteFn         func(context.Context, int) error
	updatePasswordFn func(context.Context, int, string) error
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string, includePassword bool) (*domain.User, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email, includePassword)
	}
	return nil, nil
}

func (f *fakeUserRepository) Create(ctx context.Context, user domain.User) error {
	if f.createFn != nil {
		return f.createFn(ctx, user)
	}
	return nil
}

func (f *fakeUserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	if f.getAllFn != nil {
		return f.getAllFn(ctx)
	}
	return nil, nil
}

func (f *fakeUserRepository) Update(ctx context.Context, user domain.UpdateUserCommand) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, user)
	}
	return nil
}

func (f *fakeUserRepository) Delete(ctx context.Context, id int) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

func (f *fakeUserRepository) UpdatePassword(ctx context.Context, id int, password string) error {
	if f.updatePasswordFn != nil {
		return f.updatePasswordFn(ctx, id, password)
	}
	return nil
}

type fakeAuthRepository struct {
	createVerifyEmailFn         func(context.Context, domain.VerifyEmail) error
	getVerifyEmailByIDFn        func(context.Context, string) (domain.VerifyEmail, error)
	verifyEmailFn               func(context.Context, string) error
	createForgotPasswordEmailFn func(context.Context, domain.ForgotPassword) error
	getForgotPasswordByIDFn     func(context.Context, string) (domain.ForgotPassword, error)
	deleteForgotPasswordByIDFn  func(context.Context, string) error
	createRefreshTokenFn        func(context.Context, domain.RefreshToken) error
	getRefreshTokenByHashFn     func(context.Context, []byte) (*domain.RefreshToken, error)
	useRefreshTokenFn           func(context.Context, uuid.UUID) error
	revokeRefreshTokenFn        func(context.Context, uuid.UUID) error
	revokeRefreshTokenFamilyFn  func(context.Context, uuid.UUID) error
}

func (f *fakeAuthRepository) CreateVerifyEmail(ctx context.Context, data domain.VerifyEmail) error {
	if f.createVerifyEmailFn != nil {
		return f.createVerifyEmailFn(ctx, data)
	}
	return nil
}

func (f *fakeAuthRepository) GetVerifyEmailByID(ctx context.Context, id string) (domain.VerifyEmail, error) {
	if f.getVerifyEmailByIDFn != nil {
		return f.getVerifyEmailByIDFn(ctx, id)
	}
	return domain.VerifyEmail{}, nil
}

func (f *fakeAuthRepository) VerifyEmail(ctx context.Context, id string) error {
	if f.verifyEmailFn != nil {
		return f.verifyEmailFn(ctx, id)
	}
	return nil
}

func (f *fakeAuthRepository) CreateForgotPasswordEmail(ctx context.Context, data domain.ForgotPassword) error {
	if f.createForgotPasswordEmailFn != nil {
		return f.createForgotPasswordEmailFn(ctx, data)
	}
	return nil
}

func (f *fakeAuthRepository) GetForgotPasswordByID(ctx context.Context, id string) (domain.ForgotPassword, error) {
	if f.getForgotPasswordByIDFn != nil {
		return f.getForgotPasswordByIDFn(ctx, id)
	}
	return domain.ForgotPassword{}, nil
}

func (f *fakeAuthRepository) DeleteForgotPasswordByID(ctx context.Context, id string) error {
	if f.deleteForgotPasswordByIDFn != nil {
		return f.deleteForgotPasswordByIDFn(ctx, id)
	}
	return nil
}

func (f *fakeAuthRepository) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	if f.createRefreshTokenFn != nil {
		return f.createRefreshTokenFn(ctx, token)
	}
	return nil
}

func (f *fakeAuthRepository) GetRefreshTokenByHash(ctx context.Context, hash []byte) (*domain.RefreshToken, error) {
	if f.getRefreshTokenByHashFn != nil {
		return f.getRefreshTokenByHashFn(ctx, hash)
	}
	return nil, nil
}

func (f *fakeAuthRepository) UseRefreshToken(ctx context.Context, id uuid.UUID) error {
	if f.useRefreshTokenFn != nil {
		return f.useRefreshTokenFn(ctx, id)
	}
	return nil
}

func (f *fakeAuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	if f.revokeRefreshTokenFn != nil {
		return f.revokeRefreshTokenFn(ctx, id)
	}
	return nil
}

func (f *fakeAuthRepository) RevokeRefreshTokenFamily(ctx context.Context, id uuid.UUID) error {
	if f.revokeRefreshTokenFamilyFn != nil {
		return f.revokeRefreshTokenFamilyFn(ctx, id)
	}
	return nil
}

type fakeUnitOfWork struct {
	users port.UserRepository
	auth  port.AuthRepository
}

func (f *fakeUnitOfWork) Users() port.UserRepository { return f.users }
func (f *fakeUnitOfWork) Auth() port.AuthRepository  { return f.auth }

type fakeRepository struct {
	users       port.UserRepository
	auth        port.AuthRepository
	tx          port.UnitOfWork
	withTxErr   error
	withTxFn    func(port.UnitOfWork) error
	withTxCalls int
}

func (f *fakeRepository) Users() port.UserRepository { return f.users }
func (f *fakeRepository) Auth() port.AuthRepository  { return f.auth }
func (f *fakeRepository) WithTx(_ context.Context, fn func(port.UnitOfWork) error) error {
	f.withTxCalls++
	if f.withTxFn != nil {
		return f.withTxFn(f.tx)
	}
	if f.withTxErr != nil {
		return f.withTxErr
	}
	return fn(f.tx)
}

type fakeHasher struct {
	hashFn    func(string) (string, error)
	compareFn func(string, string) error
}

func (f fakeHasher) Hash(password string) (string, error) {
	if f.hashFn != nil {
		return f.hashFn(password)
	}
	return "hashed:" + password, nil
}

func (f fakeHasher) Compare(hashed, plain string) error {
	if f.compareFn != nil {
		return f.compareFn(hashed, plain)
	}
	return nil
}

type fakeTokens struct {
	accessToken        string
	refreshToken       string
	refreshHash        []byte
	accessErr          error
	generateRefreshErr error
	hashErr            error
}

func (f fakeTokens) GenerateAccessToken(domain.User) (string, string, error) {
	return f.accessToken, "jti", f.accessErr
}

func (f fakeTokens) ValidateAccessToken(string) (map[string]any, error) {
	return nil, nil
}

func (f fakeTokens) GenerateRefreshToken() (string, []byte, error) {
	return f.refreshToken, f.refreshHash, f.generateRefreshErr
}

func (f fakeTokens) HashRefreshToken(string) ([]byte, error) {
	return f.refreshHash, f.hashErr
}

type fakeEmailSender struct {
	err   error
	calls int
}

func (f *fakeEmailSender) Send(string, string, string) error {
	f.calls++
	return f.err
}

func newAuthFakes() (*fakeRepository, *fakeUserRepository, *fakeAuthRepository, *fakeEmailSender) {
	users := &fakeUserRepository{}
	auth := &fakeAuthRepository{}
	uow := &fakeUnitOfWork{users: users, auth: auth}
	repo := &fakeRepository{users: users, auth: auth, tx: uow}
	return repo, users, auth, &fakeEmailSender{}
}

type testRateLimitStore struct {
	decision   port.RateLimitDecision
	allowErr   error
	resetErr   error
	policies   []string
	resetCalls int
}

func (s *testRateLimitStore) Allow(_ context.Context, _ string, policy port.RateLimitPolicy) (port.RateLimitDecision, error) {
	s.policies = append(s.policies, policy.Name)
	if s.allowErr != nil {
		return port.RateLimitDecision{}, s.allowErr
	}
	if s.decision.Allowed || s.decision.RetryAfter > 0 {
		return s.decision, nil
	}
	return port.RateLimitDecision{Allowed: true}, nil
}
func (s *testRateLimitStore) Reset(context.Context, string) error {
	s.resetCalls++
	return s.resetErr
}
func (*testRateLimitStore) Backend() string { return "test" }
func (*testRateLimitStore) Close() error    { return nil }

func testRateLimiterWithStore(store port.RateLimitStore) *ratelimit.RateLimiter {
	policies := make(map[string]port.RateLimitPolicy)
	for _, name := range []string{"register_email", "login_account", "verify_email", "recovery_email", "recovery_token"} {
		policies[name] = port.RateLimitPolicy{Name: name, Limit: 5, Window: time.Minute}
	}
	return ratelimit.NewRateLimiter(store, "test-key", policies)
}

func testRateLimiter() *ratelimit.RateLimiter {
	return testRateLimiterWithStore(&testRateLimitStore{decision: port.RateLimitDecision{Allowed: true}})
}

func newTestAuthService(repo port.Repository, tokens port.TokenService, email port.EmailSender, hasher port.PasswordHasher, baseURL string) *authService {
	return NewAuthService(repo, tokens, email, hasher, baseURL, testRateLimiter())
}

func TestUserService(t *testing.T) {
	user := &domain.User{ID: 7, Email: "user@example.com"}
	t.Run("gets users", func(t *testing.T) {
		_, users, _, _ := newAuthFakes()
		users.getByIDFn = func(context.Context, int) (*domain.User, error) { return user, nil }
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return user, nil }
		users.getAllFn = func(context.Context) ([]*domain.User, error) { return []*domain.User{user}, nil }
		svc := NewUserService(users)
		if got, err := svc.GetByID(context.Background(), user.ID); err != nil || got != user {
			t.Fatalf("GetByID = %#v, %v", got, err)
		}
		if got, err := svc.GetByEmail(context.Background(), user.Email); err != nil || got != user {
			t.Fatalf("GetByEmail = %#v, %v", got, err)
		}
		if got, err := svc.GetAll(context.Background()); err != nil || len(got) != 1 {
			t.Fatalf("GetAll = %#v, %v", got, err)
		}
	})

	t.Run("returns lookup errors and not found", func(t *testing.T) {
		_, users, _, _ := newAuthFakes()
		users.getByIDFn = func(context.Context, int) (*domain.User, error) { return nil, errFake }
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, errFake }
		svc := NewUserService(users)
		if _, err := svc.GetByID(context.Background(), 1); !errors.Is(err, errFake) {
			t.Fatalf("GetByID error = %v", err)
		}
		if _, err := svc.GetByEmail(context.Background(), "x"); !errors.Is(err, errFake) {
			t.Fatalf("GetByEmail error = %v", err)
		}
		users.getByIDFn = func(context.Context, int) (*domain.User, error) { return nil, nil }
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, nil }
		if _, err := svc.GetByID(context.Background(), 1); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("missing ID error = %v", err)
		}
		if _, err := svc.GetByEmail(context.Background(), "x"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("missing email error = %v", err)
		}
	})

	t.Run("handles list and mutations", func(t *testing.T) {
		_, users, _, _ := newAuthFakes()
		svc := NewUserService(users)
		users.getAllFn = func(context.Context) ([]*domain.User, error) { return nil, errFake }
		if _, err := svc.GetAll(context.Background()); !errors.Is(err, errFake) {
			t.Fatalf("GetAll error = %v", err)
		}
		command := domain.UpdateUserCommand{ID: user.ID}
		if err := svc.Update(context.Background(), 8, command); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("forbidden update = %v", err)
		}
		users.updateFn = func(context.Context, domain.UpdateUserCommand) error { return errFake }
		if err := svc.Update(context.Background(), user.ID, command); !errors.Is(err, errFake) {
			t.Fatalf("update error = %v", err)
		}
		users.updateFn = nil
		if err := svc.Update(context.Background(), user.ID, command); err != nil {
			t.Fatalf("update = %v", err)
		}
		if err := svc.Delete(context.Background(), 8, user.ID); !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("forbidden delete = %v", err)
		}
		users.deleteFn = func(context.Context, int) error { return errFake }
		if err := svc.Delete(context.Background(), user.ID, user.ID); !errors.Is(err, errFake) {
			t.Fatalf("delete error = %v", err)
		}
	})
}

func TestOAuthService(t *testing.T) {
	_, users, _, _ := newAuthFakes()
	profile := domain.OAuthProfile{UserID: "oauth-id", Email: "oauth@example.com", Name: "OAuth", Provider: "github"}

	validationTests := []struct {
		name    string
		profile domain.OAuthProfile
	}{
		{"missing user ID", domain.OAuthProfile{Email: profile.Email, Provider: profile.Provider}},
		{"missing email", domain.OAuthProfile{UserID: profile.UserID, Provider: profile.Provider}},
		{"missing provider", domain.OAuthProfile{UserID: profile.UserID, Email: profile.Email}},
	}
	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			var createCalled, getByEmailCalled bool
			users := &fakeUserRepository{
				createFn: func(context.Context, domain.User) error {
					createCalled = true
					return nil
				},
				getByEmailFn: func(context.Context, string, bool) (*domain.User, error) {
					getByEmailCalled = true
					return nil, nil
				},
			}
			if _, err := NewOAuthService(users).Login(context.Background(), tt.profile); !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("Login error = %v, want invalid input", err)
			}
			if createCalled || getByEmailCalled {
				t.Fatalf("repository called: create=%v, getByEmail=%v", createCalled, getByEmailCalled)
			}
		})
	}

	svc := NewOAuthService(users)
	users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
		return &domain.User{ID: 1, Email: profile.Email}, nil
	}
	if user, err := svc.Login(context.Background(), profile); err != nil || user.ID != 1 {
		t.Fatalf("Login = %#v, %v", user, err)
	}

	tests := []struct {
		name  string
		setup func()
	}{
		{"create error", func() {
			users.createFn = func(context.Context, domain.User) error { return errFake }
		}},
		{"get error", func() {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, errFake }
		}},
		{"not found", func() {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, nil }
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users.createFn = nil
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
				return &domain.User{ID: 1, Email: profile.Email}, nil
			}
			tt.setup()
			if _, err := svc.Login(context.Background(), profile); err == nil {
				t.Fatal("Login succeeded unexpectedly")
			}
		})
	}

	users.createFn = func(context.Context, domain.User) error { return domain.ErrConflict }
	users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
		return &domain.User{ID: 2, Email: profile.Email}, nil
	}
	if _, err := svc.Login(context.Background(), profile); err != nil {
		t.Fatalf("conflict create should continue: %v", err)
	}
}

func TestAuthServiceAppliesEachAccountPolicy(t *testing.T) {
	cases := []struct {
		name   string
		policy string
		call   func(*authService) error
	}{
		{"registration email", "register_email", func(s *authService) error {
			return s.Register(context.Background(), port.RegisterCommand{Email: "user@example.com", Password: "secret"})
		}},
		{"login account", "login_account", func(s *authService) error {
			_, _, err := s.Login(context.Background(), port.LoginCommand{Email: "user@example.com", Password: "secret"})
			return err
		}},
		{"verification email", "verify_email", func(s *authService) error {
			return s.CreateVerifyEmail(context.Background(), "user@example.com")
		}},
		{"recovery email", "recovery_email", func(s *authService) error {
			return s.ForgotPassword(context.Background(), "user@example.com")
		}},
		{"recovery token", "recovery_token", func(s *authService) error {
			return s.ResetPassword(context.Background(), uuid.NewString(), "secret")
		}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store := &testRateLimitStore{decision: port.RateLimitDecision{RetryAfter: time.Minute}}
			repo, _, _, email := newAuthFakes()
			svc := NewAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost", testRateLimiterWithStore(store))
			err := tt.call(svc)
			var limitErr *RateLimitError
			if !errors.As(err, &limitErr) || limitErr.Unavailable || limitErr.Policy != tt.policy {
				t.Fatalf("error = %#v, want denied %s policy", err, tt.policy)
			}
			if len(store.policies) != 1 || store.policies[0] != tt.policy || repo.withTxCalls != 0 {
				t.Fatalf("checked policies=%v transactions=%d, want only %s before repository work", store.policies, repo.withTxCalls, tt.policy)
			}
		})
	}
}

func TestAuthServiceRateLimitsAtApplicationBoundary(t *testing.T) {
	t.Run("registration is rejected before hashing or persistence", func(t *testing.T) {
		store := &testRateLimitStore{decision: port.RateLimitDecision{RetryAfter: 12 * time.Second}}
		repo, _, _, email := newAuthFakes()
		svc := NewAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost", testRateLimiterWithStore(store))

		err := svc.Register(context.Background(), port.RegisterCommand{Email: "user@example.com", Password: "secret"})
		var limitErr *RateLimitError
		if !errors.As(err, &limitErr) || limitErr.Unavailable || limitErr.Policy != "register_email" || limitErr.RetryAfter != 12*time.Second {
			t.Fatalf("Register error = %#v, want register_email rate-limit error", err)
		}
		if repo.withTxCalls != 0 {
			t.Fatal("rate-limited registration reached persistence")
		}
		if len(store.policies) != 1 || store.policies[0] != "register_email" {
			t.Fatalf("checked policies = %v, want [register_email]", store.policies)
		}
	})

	t.Run("successful login resets its account counter", func(t *testing.T) {
		store := &testRateLimitStore{decision: port.RateLimitDecision{Allowed: true}}
		repo, users, _, email := newAuthFakes()
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
			return &domain.User{ID: 9, Email: "user@example.com", PasswordHash: "hashed"}, nil
		}
		svc := NewAuthService(repo, &fakeTokens{accessToken: "access", refreshToken: "refresh", refreshHash: []byte("hash")}, email, fakeHasher{}, "http://localhost", testRateLimiterWithStore(store))

		if _, _, err := svc.Login(context.Background(), port.LoginCommand{Email: " USER@example.com ", Password: "secret"}); err != nil {
			t.Fatalf("Login = %v", err)
		}
		if len(store.policies) != 1 || store.policies[0] != "login_account" || store.resetCalls != 1 {
			t.Fatalf("policies=%v reset calls=%d, want login_account and one reset", store.policies, store.resetCalls)
		}
	})

	t.Run("backend failure denies login before account lookup", func(t *testing.T) {
		store := &testRateLimitStore{allowErr: errors.New("redis unavailable")}
		repo, users, _, email := newAuthFakes()
		lookupCalls := 0
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
			lookupCalls++
			return nil, nil
		}
		svc := NewAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost", testRateLimiterWithStore(store))

		_, _, err := svc.Login(context.Background(), port.LoginCommand{Email: "user@example.com", Password: "secret"})
		var limitErr *RateLimitError
		if !errors.As(err, &limitErr) || !limitErr.Unavailable || !errors.Is(err, port.ErrRateLimitBackendUnavailable) {
			t.Fatalf("Login error = %v, want backend-unavailable rate-limit error", err)
		}
		if lookupCalls != 0 {
			t.Fatalf("account lookup calls = %d, want 0", lookupCalls)
		}
	})
}

func TestAuthServiceRegisterAndEmailFlows(t *testing.T) {
	userID := 3
	user := port.RegisterCommand{Email: "user@example.com", Password: "secret"}
	t.Run("register success and failures", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(*fakeRepository, *fakeUserRepository, *fakeAuthRepository, *fakeEmailSender)
		}{
			{"hash", func(_ *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, _ *fakeEmailSender) {}},
			{"create", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeEmailSender) {
				users.createFn = func(context.Context, domain.User) error { return errFake }
			}},
			{"lookup", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeEmailSender) {
				users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, errFake }
			}},
			{"missing user", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeEmailSender) {
				users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, nil }
			}},
			{"create verification", func(_ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeEmailSender) {
				auth.createVerifyEmailFn = func(context.Context, domain.VerifyEmail) error { return errFake }
			}},
			{"send email", func(_ *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, email *fakeEmailSender) {
				email.err = errFake
			}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, users, auth, email := newAuthFakes()
				users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
					return &domain.User{ID: userID, Email: user.Email}, nil
				}
				hasher := fakeHasher{}
				if tt.name == "hash" {
					hasher.hashFn = func(string) (string, error) { return "", errFake }
				}
				tt.setup(repo, users, auth, email)
				svc := newTestAuthService(repo, fakeTokens{}, email, hasher, "http://localhost")
				if err := svc.Register(context.Background(), user); err == nil {
					t.Fatal("Register succeeded unexpectedly")
				}
			})
		}

		repo, users, _, email := newAuthFakes()
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
			return &domain.User{ID: userID, Email: user.Email}, nil
		}
		if err := newTestAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost").Register(context.Background(), user); err != nil {
			t.Fatalf("Register = %v", err)
		}
		if repo.withTxCalls != 1 {
			t.Fatalf("Register transactions = %d, want 1", repo.withTxCalls)
		}
	})

	t.Run("create verification and forgot password", func(t *testing.T) {
		repo, users, auth, email := newAuthFakes()
		users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) {
			return &domain.User{ID: userID, Email: user.Email}, nil
		}
		svc := newTestAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost")
		if err := svc.CreateVerifyEmail(context.Background(), user.Email); err != nil {
			t.Fatalf("CreateVerifyEmail = %v", err)
		}
		if err := svc.ForgotPassword(context.Background(), user.Email); err != nil {
			t.Fatalf("ForgotPassword = %v", err)
		}
		email.err = errFake
		if err := svc.CreateVerifyEmail(context.Background(), user.Email); err == nil {
			t.Fatal("CreateVerifyEmail succeeded with email failure")
		}
		email.err = nil
		auth.createForgotPasswordEmailFn = func(context.Context, domain.ForgotPassword) error { return errFake }
		if err := svc.ForgotPassword(context.Background(), user.Email); err == nil {
			t.Fatal("ForgotPassword succeeded with repository failure")
		}
	})
}

func TestAuthServiceLogin(t *testing.T) {
	baseUser := &domain.User{ID: 4, Email: "user@example.com", PasswordHash: "hashed"}
	tests := []struct {
		name  string
		setup func(*fakeRepository, *fakeUserRepository, *fakeAuthRepository, *fakeHasher, *fakeTokens)
	}{
		{"lookup error", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeHasher, _ *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, errFake }
		}},
		{"not found", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeHasher, _ *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return nil, nil }
		}},
		{"bad password", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, hasher *fakeHasher, _ *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return baseUser, nil }
			hasher.compareFn = func(string, string) error { return errFake }
		}},
		{"access token", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeHasher, tokens *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return baseUser, nil }
			tokens.accessErr = errFake
		}},
		{"refresh token", func(_ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeHasher, tokens *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return baseUser, nil }
			tokens.generateRefreshErr = errFake
		}},
		{"persist token", func(_ *fakeRepository, users *fakeUserRepository, auth *fakeAuthRepository, _ *fakeHasher, _ *fakeTokens) {
			users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return baseUser, nil }
			auth.createRefreshTokenFn = func(context.Context, domain.RefreshToken) error { return errFake }
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, users, auth, email := newAuthFakes()
			hasher := &fakeHasher{}
			tokens := &fakeTokens{accessToken: "access", refreshToken: "refresh", refreshHash: []byte("hash")}
			tt.setup(repo, users, auth, hasher, tokens)
			_, _, err := newTestAuthService(repo, tokens, email, hasher, "http://localhost").Login(context.Background(), port.LoginCommand{Email: baseUser.Email, Password: "password"})
			if err == nil {
				t.Fatal("Login succeeded unexpectedly")
			}
		})
	}

	repo, users, auth, email := newAuthFakes()
	users.getByEmailFn = func(context.Context, string, bool) (*domain.User, error) { return baseUser, nil }
	tokens := &fakeTokens{accessToken: "access", refreshToken: "refresh", refreshHash: []byte("hash")}
	access, refresh, err := newTestAuthService(repo, tokens, email, fakeHasher{}, "http://localhost").Login(context.Background(), port.LoginCommand{Email: baseUser.Email, Password: "password"})
	if err != nil || access != "access" || refresh != "refresh" {
		t.Fatalf("Login = %q, %q, %v", access, refresh, err)
	}
	_ = auth
}

func TestAuthServiceVerifyAndReset(t *testing.T) {
	validID := uuid.New().String()
	userID := 5
	t.Run("verify email", func(t *testing.T) {
		repo, _, auth, email := newAuthFakes()
		svc := newTestAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost")
		if err := svc.VerifyEmail(context.Background(), "bad"); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("invalid ID error = %v", err)
		}
		auth.getVerifyEmailByIDFn = func(context.Context, string) (domain.VerifyEmail, error) { return domain.VerifyEmail{}, errFake }
		if err := svc.VerifyEmail(context.Background(), validID); !errors.Is(err, errFake) {
			t.Fatalf("lookup error = %v", err)
		}
		auth.getVerifyEmailByIDFn = func(context.Context, string) (domain.VerifyEmail, error) { return domain.VerifyEmail{}, nil }
		if err := svc.VerifyEmail(context.Background(), validID); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("missing error = %v", err)
		}
		auth.getVerifyEmailByIDFn = func(context.Context, string) (domain.VerifyEmail, error) {
			return domain.VerifyEmail{ID: uuid.New(), ExpiredAt: time.Now().Add(-time.Minute)}, nil
		}
		if err := svc.VerifyEmail(context.Background(), validID); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expired error = %v", err)
		}
		auth.getVerifyEmailByIDFn = func(context.Context, string) (domain.VerifyEmail, error) {
			return domain.VerifyEmail{ID: uuid.New(), ExpiredAt: time.Now().Add(time.Minute)}, nil
		}
		auth.verifyEmailFn = func(context.Context, string) error { return errFake }
		if err := svc.VerifyEmail(context.Background(), validID); !errors.Is(err, errFake) {
			t.Fatalf("verify error = %v", err)
		}
		auth.verifyEmailFn = nil
		if err := svc.VerifyEmail(context.Background(), validID); err != nil {
			t.Fatalf("verify = %v", err)
		}
	})

	t.Run("reset password", func(t *testing.T) {
		repo, users, auth, email := newAuthFakes()
		svc := newTestAuthService(repo, fakeTokens{}, email, fakeHasher{}, "http://localhost")
		if err := svc.ResetPassword(context.Background(), "bad", "new"); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("invalid ID error = %v", err)
		}
		hasher := fakeHasher{hashFn: func(string) (string, error) { return "", errFake }}
		if err := newTestAuthService(repo, fakeTokens{}, email, hasher, "http://localhost").ResetPassword(context.Background(), validID, "new"); !errors.Is(err, errFake) {
			t.Fatalf("hash error = %v", err)
		}
		auth.getForgotPasswordByIDFn = func(context.Context, string) (domain.ForgotPassword, error) { return domain.ForgotPassword{}, errFake }
		if err := svc.ResetPassword(context.Background(), validID, "new"); !errors.Is(err, errFake) {
			t.Fatalf("lookup error = %v", err)
		}
		auth.getForgotPasswordByIDFn = func(context.Context, string) (domain.ForgotPassword, error) { return domain.ForgotPassword{}, nil }
		if err := svc.ResetPassword(context.Background(), validID, "new"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("missing error = %v", err)
		}
		auth.getForgotPasswordByIDFn = func(context.Context, string) (domain.ForgotPassword, error) {
			return domain.ForgotPassword{ID: uuid.New(), UserID: userID, ExpiredAt: time.Now().Add(-time.Minute)}, nil
		}
		if err := svc.ResetPassword(context.Background(), validID, "new"); !errors.Is(err, domain.ErrInvalidInput) {
			t.Fatalf("expired error = %v", err)
		}
		auth.getForgotPasswordByIDFn = func(context.Context, string) (domain.ForgotPassword, error) {
			return domain.ForgotPassword{ID: uuid.New(), UserID: userID, ExpiredAt: time.Now().Add(time.Minute)}, nil
		}
		auth.deleteForgotPasswordByIDFn = func(context.Context, string) error { return errFake }
		if err := svc.ResetPassword(context.Background(), validID, "new"); !errors.Is(err, errFake) {
			t.Fatalf("delete error = %v", err)
		}
		auth.deleteForgotPasswordByIDFn = nil
		users.updatePasswordFn = func(context.Context, int, string) error { return errFake }
		if err := svc.ResetPassword(context.Background(), validID, "new"); !errors.Is(err, errFake) {
			t.Fatalf("update error = %v", err)
		}
		users.updatePasswordFn = nil
		if err := svc.ResetPassword(context.Background(), validID, "new"); err != nil {
			t.Fatalf("reset = %v", err)
		}
	})
}

func TestAuthServiceRefreshAndLogout(t *testing.T) {
	tokenID := uuid.New()
	familyID := uuid.New()
	baseToken := &domain.RefreshToken{ID: tokenID, UserID: 9, FamilyID: familyID, ExpiredAt: time.Now().Add(time.Hour)}
	user := &domain.User{ID: 9, Email: "user@example.com"}
	newService := func() (*authService, *fakeRepository, *fakeUserRepository, *fakeAuthRepository, *fakeTokens) {
		repo, users, auth, email := newAuthFakes()
		users.getByIDFn = func(context.Context, int) (*domain.User, error) { return user, nil }
		auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return baseToken, nil }
		tokens := &fakeTokens{accessToken: "access", refreshToken: "new-refresh", refreshHash: []byte("hash")}
		return newTestAuthService(repo, tokens, email, fakeHasher{}, "http://localhost"), repo, users, auth, tokens
	}

	t.Run("refresh success", func(t *testing.T) {
		svc, repo, _, _, _ := newService()
		access, refresh, err := svc.RefreshTokens(context.Background(), "old", "127.0.0.1", "agent")
		if err != nil || access != "access" || refresh != "new-refresh" {
			t.Fatalf("RefreshTokens = %q, %q, %v", access, refresh, err)
		}
		if repo.withTxCalls != 1 {
			t.Fatalf("RefreshTokens transactions = %d, want 1", repo.withTxCalls)
		}
	})

	t.Run("refresh errors", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(*authService, *fakeRepository, *fakeUserRepository, *fakeAuthRepository, *fakeTokens)
		}{
			{"hash", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, tokens *fakeTokens) {
				tokens.hashErr = errFake
			}},
			{"transaction", func(_ *authService, repo *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, _ *fakeTokens) {
				repo.withTxErr = errFake
			}},
			{"lookup", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return nil, errFake }
			}},
			{"missing", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return nil, nil }
			}},
			{"expired", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) {
					return &domain.RefreshToken{ExpiredAt: time.Now().Add(-time.Minute)}, nil
				}
			}},
			{"revoked", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) {
					now := time.Now()
					return &domain.RefreshToken{ExpiredAt: time.Now().Add(time.Hour), RevokedAt: &now}, nil
				}
			}},
			{"revoke family", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				now := time.Now()
				auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) {
					return &domain.RefreshToken{ExpiredAt: time.Now().Add(time.Hour), UsedAt: &now, FamilyID: familyID}, nil
				}
				auth.revokeRefreshTokenFamilyFn = func(context.Context, uuid.UUID) error { return errFake }
			}},
			{"use token", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.useRefreshTokenFn = func(context.Context, uuid.UUID) error { return errFake }
			}},
			{"new refresh", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, tokens *fakeTokens) {
				tokens.generateRefreshErr = errFake
			}},
			{"create new token", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, auth *fakeAuthRepository, _ *fakeTokens) {
				auth.createRefreshTokenFn = func(context.Context, domain.RefreshToken) error { return errFake }
			}},
			{"get user", func(_ *authService, _ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeTokens) {
				users.getByIDFn = func(context.Context, int) (*domain.User, error) { return nil, errFake }
			}},
			{"user missing", func(_ *authService, _ *fakeRepository, users *fakeUserRepository, _ *fakeAuthRepository, _ *fakeTokens) {
				users.getByIDFn = func(context.Context, int) (*domain.User, error) { return nil, nil }
			}},
			{"access token", func(_ *authService, _ *fakeRepository, _ *fakeUserRepository, _ *fakeAuthRepository, tokens *fakeTokens) {
				tokens.accessErr = errFake
			}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc, repo, users, auth, tokens := newService()
				tt.setup(svc, repo, users, auth, tokens)
				if _, _, err := svc.RefreshTokens(context.Background(), "old", "ip", "agent"); err == nil {
					t.Fatal("RefreshTokens succeeded unexpectedly")
				}
			})
		}
	})

	t.Run("reused token", func(t *testing.T) {
		svc, _, _, auth, _ := newService()
		now := time.Now()
		auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) {
			return &domain.RefreshToken{ExpiredAt: time.Now().Add(time.Hour), UsedAt: &now, FamilyID: familyID}, nil
		}
		if _, _, err := svc.RefreshTokens(context.Background(), "old", "ip", "agent"); !errors.Is(err, ErrRefreshTokenReused) {
			t.Fatalf("reuse error = %v", err)
		}
	})

	t.Run("logout", func(t *testing.T) {
		svc, _, _, auth, tokens := newService()
		tokens.hashErr = errFake
		if err := svc.Logout(context.Background(), "token"); !errors.Is(err, errFake) {
			t.Fatalf("hash error = %v", err)
		}
		tokens.hashErr = nil
		auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return nil, errFake }
		if err := svc.Logout(context.Background(), "token"); !errors.Is(err, errFake) {
			t.Fatalf("lookup error = %v", err)
		}
		auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return nil, nil }
		if err := svc.Logout(context.Background(), "token"); err != nil {
			t.Fatalf("missing logout = %v", err)
		}
		auth.getRefreshTokenByHashFn = func(context.Context, []byte) (*domain.RefreshToken, error) { return baseToken, nil }
		auth.revokeRefreshTokenFamilyFn = func(context.Context, uuid.UUID) error { return errFake }
		if err := svc.Logout(context.Background(), "token"); !errors.Is(err, errFake) {
			t.Fatalf("revoke error = %v", err)
		}
		auth.revokeRefreshTokenFamilyFn = nil
		if err := svc.Logout(context.Background(), "token"); err != nil {
			t.Fatalf("logout = %v", err)
		}
	})
}

var _ port.Repository = (*fakeRepository)(nil)
var _ port.UserRepository = (*fakeUserRepository)(nil)
var _ port.AuthRepository = (*fakeAuthRepository)(nil)
var _ port.UnitOfWork = (*fakeUnitOfWork)(nil)
var _ port.TokenService = (fakeTokens{})
var _ port.PasswordHasher = (fakeHasher{})
var _ port.EmailSender = (*fakeEmailSender)(nil)
