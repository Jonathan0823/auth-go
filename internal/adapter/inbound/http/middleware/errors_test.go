package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"invalid input", fmt.Errorf("bad field: %w", domain.ErrInvalidInput), http.StatusBadRequest, "invalid input"},
		{"unauthenticated", domain.ErrUnauthenticated, http.StatusUnauthorized, "unauthorized"},
		{"forbidden", domain.ErrForbidden, http.StatusForbidden, "forbidden"},
		{"not found", domain.ErrNotFound, http.StatusNotFound, "not found"},
		{"conflict", domain.ErrConflict, http.StatusConflict, "conflict"},
		{"unknown", errors.New("database password leaked"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message := mapError(tt.err)
			if status != tt.status || message != tt.message {
				t.Fatalf("mapError() = (%d, %q), want (%d, %q)", status, message, tt.status, tt.message)
			}
		})
	}
}
