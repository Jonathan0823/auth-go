//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	outpostgres "github.com/Jonathan0823/auth-go/internal/adapter/outbound/postgres"
	"github.com/Jonathan0823/auth-go/internal/core/domain"
	"github.com/Jonathan0823/auth-go/internal/core/port"
)

func connectPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestCreateAndGetRefreshToken(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()

	userID := setupUser(t, pool)

	now := time.Now().UTC()
	familyID := uuid.New()
	hash := []byte("test-hmac-hash-1-32bytes-of-padding!!!")
	rt := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hash,
		FamilyID:  familyID,
		ParentID:  nil,
		ExpiredAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	if err := repo.Auth().CreateRefreshToken(ctx, rt); err != nil {
		t.Fatalf("CreateRefreshToken: %v", err)
	}

	got, err := repo.Auth().GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash: %v", err)
	}
	if got == nil {
		t.Fatal("GetRefreshTokenByHash: nil result")
	}
	if got.ID != rt.ID {
		t.Fatalf("ID: got %v, want %v", got.ID, rt.ID)
	}
	if got.UserID != userID {
		t.Fatalf("UserID: got %d, want %d", got.UserID, userID)
	}
	if got.FamilyID != familyID {
		t.Fatalf("FamilyID: got %v, want %v", got.FamilyID, familyID)
	}
	if got.UsedAt != nil {
		t.Fatal("UsedAt should be nil for new token")
	}
	if got.RevokedAt != nil {
		t.Fatal("RevokedAt should be nil for new token")
	}
}

func TestUseRefreshToken(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()

	userID := setupUser(t, pool)
	rt := createTestToken(t, repo, userID)

	if err := repo.Auth().UseRefreshToken(ctx, rt.ID); err != nil {
		t.Fatalf("UseRefreshToken: %v", err)
	}

	got, err := repo.Auth().GetRefreshTokenByHash(ctx, rt.TokenHash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash: %v", err)
	}
	if got == nil {
		t.Fatal("GetRefreshTokenByHash: nil result")
	}
	if got.UsedAt == nil || got.UsedAt.IsZero() {
		t.Fatal("UsedAt should be set after UseRefreshToken")
	}
}

func TestRevokeRefreshToken(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()

	userID := setupUser(t, pool)
	rt := createTestToken(t, repo, userID)

	if err := repo.Auth().RevokeRefreshToken(ctx, rt.ID); err != nil {
		t.Fatalf("RevokeRefreshToken: %v", err)
	}

	got, err := repo.Auth().GetRefreshTokenByHash(ctx, rt.TokenHash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash: %v", err)
	}
	if got == nil {
		t.Fatal("GetRefreshTokenByHash: nil result")
	}
	if got.RevokedAt == nil || got.RevokedAt.IsZero() {
		t.Fatal("RevokedAt should be set after RevokeRefreshToken")
	}
}

func TestRevokeRefreshTokenFamily(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()

	userID := setupUser(t, pool)
	familyID := uuid.New()

	t1 := createTestTokenInFamily(t, repo, userID, familyID, nil)
	t2 := createTestTokenInFamily(t, repo, userID, familyID, &t1.ID)

	if err := repo.Auth().RevokeRefreshTokenFamily(ctx, familyID); err != nil {
		t.Fatalf("RevokeRefreshTokenFamily: %v", err)
	}

	got1, err := repo.Auth().GetRefreshTokenByHash(ctx, t1.TokenHash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash t1: %v", err)
	}
	if got1 == nil {
		t.Fatal("GetRefreshTokenByHash t1: nil")
	}
	if got1.RevokedAt == nil || got1.RevokedAt.IsZero() {
		t.Fatal("t1 should be revoked")
	}

	got2, err := repo.Auth().GetRefreshTokenByHash(ctx, t2.TokenHash)
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash t2: %v", err)
	}
	if got2 == nil {
		t.Fatal("GetRefreshTokenByHash t2: nil")
	}
	if got2.RevokedAt == nil || got2.RevokedAt.IsZero() {
		t.Fatal("t2 should be revoked")
	}
}

func TestGetRefreshTokenByHash_NotFound(t *testing.T) {
	pool := connectPool(t)
	repo := outpostgres.NewRepository(pool)
	ctx := context.Background()

	got, err := repo.Auth().GetRefreshTokenByHash(ctx, []byte("nonexistent-hash"))
	if err != nil {
		t.Fatalf("GetRefreshTokenByHash: %v", err)
	}
	if got != nil {
		t.Fatal("Expected nil for nonexistent hash")
	}
}

// --

var userCounter int

func setupUser(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	userCounter++
	username := "test-refresh-" + uuid.New().String()[:8]
	email := username + "@test.com"
	var id int32
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id`,
		username, email, "test-password-argon2id",
	).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return int(id)
}

func createTestToken(t *testing.T, repo port.Repository, userID int) domain.RefreshToken {
	t.Helper()
	now := time.Now().UTC()
	uid := uuid.New()
	rt := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: uid[:],
		FamilyID:  uuid.New(),
		ParentID:  nil,
		ExpiredAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	}
	if err := repo.Auth().CreateRefreshToken(context.Background(), rt); err != nil {
		t.Fatalf("createTestToken: %v", err)
	}
	return rt
}

func createTestTokenInFamily(t *testing.T, repo port.Repository, userID int, familyID uuid.UUID, parentID *uuid.UUID) domain.RefreshToken {
	t.Helper()
	now := time.Now().UTC()
	uid := uuid.New()
	rt := domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: uid[:],
		FamilyID:  familyID,
		ParentID:  parentID,
		ExpiredAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		IPAddress: "127.0.0.1",
		UserAgent: "test",
	}
	if err := repo.Auth().CreateRefreshToken(context.Background(), rt); err != nil {
		t.Fatalf("createTestTokenInFamily: %v", err)
	}
	return rt
}
