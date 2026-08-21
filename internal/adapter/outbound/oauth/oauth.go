package oauth

import (
	"net/http"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
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
	SecureCookies      bool
}

type Client struct{}

func New(cfg Config) *Client {
	configure(cfg)
	return &Client{}
}

func configure(cfg Config) {
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = cfg.SecureCookies
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store
	providers := make([]goth.Provider, 0, 2)
	if cfg.GitHubClientID != "" {
		providers = append(providers, github.New(
			cfg.GitHubClientID,
			cfg.GitHubClientSecret,
			cfg.BaseURL+"/api/oauth/github/callback",
			"user:email",
		))
	}
	if cfg.GoogleClientID != "" {
		providers = append(providers, google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.BaseURL+"/api/oauth/google/callback",
		))
	}
	goth.ClearProviders()
	goth.UseProviders(providers...)
}

func (c *Client) BeginAuth(w http.ResponseWriter, r *http.Request, provider string) {
	withProvider(provider, func() {
		gothic.BeginAuthHandler(w, r)
	})
}

func (c *Client) CompleteAuth(w http.ResponseWriter, r *http.Request, provider string) (domain.OAuthProfile, error) {
	var user goth.User
	err := withProviderErr(provider, func() error {
		var err error
		user, err = gothic.CompleteUserAuth(w, r)
		return err
	})
	if err != nil {
		return domain.OAuthProfile{}, err
	}
	return profileFromUser(user), nil
}

func profileFromUser(user goth.User) domain.OAuthProfile {
	return domain.OAuthProfile{
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
