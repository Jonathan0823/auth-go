package oauth

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

func TestProfileFromUser(t *testing.T) {
	got := profileFromUser(goth.User{
		UserID:    "123",
		Email:     "user@example.com",
		NickName:  "nick",
		Name:      "name",
		Provider:  "github",
		AvatarURL: "https://example.com/avatar.png",
	})

	if got.UserID != "123" || got.Email != "user@example.com" || got.Name != "nick" || got.Provider != "github" || got.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "", "fallback"); got != "fallback" {
		t.Fatalf("firstNonEmpty() = %q, want %q", got, "fallback")
	}
	if got := firstNonEmpty("primary", "fallback"); got != "primary" {
		t.Fatalf("firstNonEmpty() = %q, want %q", got, "primary")
	}
}

func TestNewAndProviderHooks(t *testing.T) {
	New(Config{
		BaseURL:            "https://example.com",
		SessionSecret:      "session-secret",
		GitHubClientID:     "github-id",
		GitHubClientSecret: "github-secret",
		SecureCookies:      true,
	})
	store, ok := gothic.Store.(*sessions.CookieStore)
	if !ok || !store.Options.Secure {
		t.Fatal("OAuth session cookie is not secure")
	}
	provider, err := goth.GetProvider("github")
	if err != nil {
		t.Fatalf("GitHub provider is not configured: %v", err)
	}
	session, err := provider.BeginAuth("state")
	if err != nil {
		t.Fatalf("BeginAuth() error = %v", err)
	}
	authURL, err := session.GetAuthURL()
	if err != nil || !strings.Contains(authURL, "%2Fapi%2Foauth%2Fgithub%2Fcallback") {
		t.Fatalf("OAuth callback URL = %q, error = %v", authURL, err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	withProvider("github", func() {
		provider, err := gothic.GetProviderName(req)
		if err != nil || provider != "github" {
			t.Fatalf("provider = %q, err = %v", provider, err)
		}
	})
	if err := withProviderErr("google", func() error { return errors.New("provider failure") }); err == nil {
		t.Fatal("withProviderErr succeeded unexpectedly")
	}
}
