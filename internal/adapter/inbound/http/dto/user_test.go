package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func TestUserResponseMappings(t *testing.T) {
	now := time.Now()
	user := &domain.User{
		ID:         1,
		OAuthID:    "oauth-id",
		Username:   "user",
		AvatarURL:  "https://example.com/avatar.png",
		Email:      "user@example.com",
		IsVerified: true,
		Provider:   "github",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	user.PasswordHash = "must-not-be-exposed"
	response := UserResponseFromDomain(user)
	if response.ID != user.ID || response.Email != user.Email || response.Provider != user.Provider || !response.IsVerified {
		t.Fatalf("response = %#v", response)
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "must-not-be-exposed") || strings.Contains(string(encoded), "password") {
		t.Fatalf("user response exposed password data: %s", encoded)
	}
	responses := UserResponsesFromDomain([]*domain.User{user})
	if len(responses) != 1 || responses[0].ID != user.ID {
		t.Fatalf("responses = %#v", responses)
	}
}
