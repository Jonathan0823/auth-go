package oauth

import (
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"

	"github.com/Jonathan0823/auth-go/internal/core/port"
)

// ponytail: global lock around gothic's provider hook; switch to provider-scoped auth if throughput matters.
var gothicMu sync.Mutex

type Config struct {
	BaseURL            string
	SessionSecret      string
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
}

type client struct{}

func New(cfg Config) port.OAuthClient {
	configure(cfg)
	return &client{}
}

func configure(cfg Config) {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store
	goth.UseProviders(
		github.New(
			cfg.GitHubClientID,
			cfg.GitHubClientSecret,
			cfg.BaseURL+"/api/auth/github/callback",
			"user:email",
		),
		google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.BaseURL+"/api/auth/google/callback",
		),
	)
}

func (c *client) BeginAuth(w http.ResponseWriter, r *http.Request, provider string) {
	withProvider(provider, func() {
		gothic.BeginAuthHandler(w, r)
	})
}

func (c *client) CompleteAuth(w http.ResponseWriter, r *http.Request, provider string) (port.OAuthProfile, error) {
	var user goth.User
	err := withProviderErr(provider, func() error {
		var err error
		user, err = gothic.CompleteUserAuth(w, r)
		return err
	})
	if err != nil {
		return port.OAuthProfile{}, err
	}
	return profileFromUser(user), nil
}

func profileFromUser(user goth.User) port.OAuthProfile {
	return port.OAuthProfile{
		UserID:    user.UserID,
		Email:     user.Email,
		Name:      firstNonEmpty(user.NickName, user.Name),
		Provider:  user.Provider,
		AvatarURL: user.AvatarURL,
	}
}

func withProvider(provider string, fn func()) {
	gothicMu.Lock()
	defer gothicMu.Unlock()
	prev := gothic.GetProviderName
	gothic.GetProviderName = func(*http.Request) (string, error) { return provider, nil }
	defer func() { gothic.GetProviderName = prev }()
	fn()
}

func withProviderErr(provider string, fn func() error) error {
	gothicMu.Lock()
	defer gothicMu.Unlock()
	prev := gothic.GetProviderName
	gothic.GetProviderName = func(*http.Request) (string, error) { return provider, nil }
	defer func() { gothic.GetProviderName = prev }()
	return fn()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
