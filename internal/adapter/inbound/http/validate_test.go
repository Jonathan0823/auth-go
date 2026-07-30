package http

import (
	"testing"

	"github.com/Jonathan0823/auth-go/internal/adapter/inbound/http/dto"
)

func TestValidateStruct(t *testing.T) {
	if got := ValidateStruct(dto.CredentialsRequest{
		Email:    "user@example.com",
		Password: "password123",
	}); got != nil {
		t.Fatalf("valid request returned validation errors: %v", got)
	}

	got := ValidateStruct(dto.CredentialsRequest{
		Email:    "invalid-email",
		Password: "short",
	})
	if got["email"] != "email must be a valid email" {
		t.Errorf("email error = %q", got["email"])
	}
	if got["password"] != "password must be at least 8 characters" {
		t.Errorf("password error = %q", got["password"])
	}
}
