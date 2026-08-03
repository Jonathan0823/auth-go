//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
)

func TestUserRepositoryCRUD(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()
	email := "user-repository-" + uuid.NewString() + "@example.com"
	username := "user-repository"
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE email = $1`, email)
	})

	if err := repo.Users().Create(ctx, domain.User{Username: username, Email: email, Password: "seed-value"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Users().Create(ctx, domain.User{Username: username, Email: email, Password: "duplicate-value"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate Create error = %v, want conflict", err)
	}

	storedUser, err := repo.Users().GetByEmail(ctx, email, true)
	if err != nil {
		t.Fatalf("GetByEmail with password: %v", err)
	}
	if storedUser == nil || storedUser.Password != "seed-value" {
		t.Fatalf("GetByEmail with password = %#v", storedUser)
	}
	withoutPassword, err := repo.Users().GetByEmail(ctx, email, false)
	if err != nil {
		t.Fatalf("GetByEmail without password: %v", err)
	}
	if withoutPassword == nil || withoutPassword.Password != "" {
		t.Fatalf("GetByEmail without password = %#v", withoutPassword)
	}

	byID, err := repo.Users().GetByID(ctx, storedUser.ID)
	if err != nil || byID == nil || byID.Email != email {
		t.Fatalf("GetByID = %#v, %v", byID, err)
	}
	if missing, err := repo.Users().GetByID(ctx, 999999999); err != nil || missing != nil {
		t.Fatalf("missing GetByID = %#v, %v", missing, err)
	}
	if missing, err := repo.Users().GetByEmail(ctx, "missing-"+email, false); err != nil || missing != nil {
		t.Fatalf("missing GetByEmail = %#v, %v", missing, err)
	}

	if err := repo.Users().Update(ctx, domain.UpdateUserCommand{ID: storedUser.ID, Username: "updated-user", Email: email}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if err := repo.Users().UpdatePassword(ctx, storedUser.ID, "updated-value"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}
	updated, err := repo.Users().GetByEmail(ctx, email, true)
	if err != nil || updated == nil || updated.Username != "updated-user" || updated.Password != "updated-value" {
		t.Fatalf("updated user = %#v, %v", updated, err)
	}

	all, err := repo.Users().GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	found := false
	for _, candidate := range all {
		if candidate.ID == storedUser.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("GetAll did not contain user %d", storedUser.ID)
	}

	if err := repo.Users().Delete(ctx, storedUser.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleted, err := repo.Users().GetByID(ctx, storedUser.ID); err != nil || deleted != nil {
		t.Fatalf("deleted GetByID = %#v, %v", deleted, err)
	}
}

func TestEmailTokenRepositories(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()
	userID := setupUser(t, pool)

	verifyID := uuid.New()
	if err := repo.Auth().CreateVerifyEmail(ctx, domain.VerifyEmail{
		ID:        verifyID,
		UserID:    userID,
		Email:     "verify@example.com",
		ExpiredAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateVerifyEmail: %v", err)
	}
	verify, err := repo.Auth().GetVerifyEmailByID(ctx, verifyID.String())
	if err != nil || verify.ID != verifyID || verify.Email != "verify@example.com" {
		t.Fatalf("GetVerifyEmailByID = %#v, %v", verify, err)
	}
	if err := repo.Auth().VerifyEmail(ctx, verifyID.String()); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}
	if missing, err := repo.Auth().GetVerifyEmailByID(ctx, verifyID.String()); err != nil || missing.ID != uuid.Nil {
		t.Fatalf("deleted verification = %#v, %v", missing, err)
	}

	forgotID := uuid.New()
	if err := repo.Auth().CreateForgotPasswordEmail(ctx, domain.ForgotPassword{
		ID:        forgotID,
		UserID:    userID,
		Email:     "forgot@example.com",
		ExpiredAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateForgotPasswordEmail: %v", err)
	}
	forgot, err := repo.Auth().GetForgotPasswordByID(ctx, forgotID.String())
	if err != nil || forgot.ID != forgotID || forgot.UserID != userID {
		t.Fatalf("GetForgotPasswordByID = %#v, %v", forgot, err)
	}
	if err := repo.Auth().DeleteForgotPasswordByID(ctx, forgotID.String()); err != nil {
		t.Fatalf("DeleteForgotPasswordByID: %v", err)
	}
	if missing, err := repo.Auth().GetForgotPasswordByID(ctx, forgotID.String()); err != nil || missing.ID != uuid.Nil {
		t.Fatalf("deleted forgot password = %#v, %v", missing, err)
	}
}
