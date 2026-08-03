package dto

import (
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
	response := UserResponseFromDomain(user)
	if response.ID != user.ID || response.Email != user.Email || response.Provider != user.Provider || !response.IsVerified {
		t.Fatalf("response = %#v", response)
	}
	responses := UserResponsesFromDomain([]*domain.User{user})
	if len(responses) != 1 || responses[0].ID != user.ID {
		t.Fatalf("responses = %#v", responses)
	}
}
