package oauth

import (
	"testing"

	"github.com/markbates/goth"
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
