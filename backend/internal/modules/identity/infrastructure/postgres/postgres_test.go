package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	auditDomain "pidiekab/internal/modules/audit/domain"
	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	"pidiekab/internal/modules/identity/domain"
	"pidiekab/internal/modules/identity/infrastructure/postgres"
	"pidiekab/internal/platform/database"
)

func getTestPool(t *testing.T) *database.Config {
	dbURL := os.Getenv("PIDIEKAB_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}
	return &database.Config{
		URL:             dbURL,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: 10 * time.Minute,
	}
}

func TestPostgresRepositoriesIntegration(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cfg := getTestPool(t)

	pool, err := database.NewPool(ctx, *cfg, logger)
	if err != nil {
		t.Skipf("skipping postgres integration test: database not reachable: %v", err)
		return
	}
	defer pool.Close()

	adminRepo := postgres.NewAdminRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	auditRepo := auditPostgres.NewRepository(pool)

	testEmail := "test_admin_" + time.Now().Format("20060102150405.000000") + "@pidiekab.id"

	// 1. Create Admin
	admin := &domain.Admin{
		Email:        testEmail,
		PasswordHash: "$argon2id$v=19$m=1024,t=1,p=1$fakeSalt$fakeHash",
		DisplayName:  "Test Integration Admin",
		IsActive:     true,
	}

	if err := adminRepo.Create(ctx, admin); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	if admin.ID <= 0 {
		t.Errorf("expected positive admin ID, got %d", admin.ID)
	}

	// 2. Duplicate email should return ErrEmailAlreadyExists
	dupAdmin := &domain.Admin{
		Email:        testEmail,
		PasswordHash: "any",
		DisplayName:  "Duplicate",
		IsActive:     true,
	}
	if err := adminRepo.Create(ctx, dupAdmin); err != domain.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 3. Lookup by ID and Email
	byID, err := adminRepo.GetByID(ctx, admin.ID)
	if err != nil {
		t.Fatalf("failed to get admin by id: %v", err)
	}
	if byID.Email != testEmail {
		t.Errorf("expected email %q, got %q", testEmail, byID.Email)
	}

	byEmail, err := adminRepo.GetByEmail(ctx, testEmail)
	if err != nil {
		t.Fatalf("failed to get admin by email: %v", err)
	}
	if byEmail.ID != admin.ID {
		t.Errorf("expected id %d, got %d", admin.ID, byEmail.ID)
	}

	// 4. Update password
	newHash := "$argon2id$v=19$m=1024,t=1,p=1$newSalt$newHash"
	if err := adminRepo.UpdatePassword(ctx, admin.ID, newHash); err != nil {
		t.Fatalf("failed to update password: %v", err)
	}
	updatedAdmin, _ := adminRepo.GetByID(ctx, admin.ID)
	if updatedAdmin.PasswordHash != newHash {
		t.Errorf("password hash was not updated")
	}

	// 5. Update last login
	now := time.Now().UTC()
	if err := adminRepo.UpdateLastLogin(ctx, admin.ID, now); err != nil {
		t.Fatalf("failed to update last login: %v", err)
	}

	// 6. Create Session
	tokenHash := "thash_" + time.Now().Format("20060102150405.000000")
	sess := &domain.Session{
		AdminID:   admin.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(2 * time.Hour),
	}
	if err := sessionRepo.Create(ctx, sess); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// 7. Lookup active session
	activeSess, err := sessionRepo.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != nil {
		t.Fatalf("failed to get active session: %v", err)
	}
	if activeSess.AdminID != admin.ID {
		t.Errorf("expected session admin ID %d, got %d", admin.ID, activeSess.AdminID)
	}

	// 8. Revoke session
	if err := sessionRepo.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}
	_, err = sessionRepo.GetActiveByTokenHash(ctx, tokenHash, now)
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for revoked session, got %v", err)
	}

	// 9. Record audit log
	ipStr := "127.0.0.1"
	err = auditRepo.Record(ctx, auditDomain.Log{
		AdminID:    &admin.ID,
		Action:     "test.integration",
		EntityType: "admin",
		EntityID:   &admin.ID,
		Metadata:   map[string]any{"status": "testing"},
		IPAddress:  &ipStr,
	})
	if err != nil {
		t.Fatalf("failed to record audit log: %v", err)
	}
}
