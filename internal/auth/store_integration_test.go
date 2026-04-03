// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/start-codex/tookly/internal/sessions"
	"github.com/start-codex/tookly/internal/testpg"
)

// --- user integration tests ---

func TestCreateUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (CreateParams, func(*testing.T))
		wantErr error
	}{
		{
			name: "creates user successfully",
			arrange: func(t *testing.T, db *sqlx.DB) (CreateParams, func(*testing.T)) {
				params := CreateParams{Email: uniqueEmail(t, db), Name: "Alice", Password: "pass123"}
				return params, func(t *testing.T) {}
			},
		},
		{
			name:    "duplicate email",
			wantErr: ErrDuplicateEmail,
			arrange: func(t *testing.T, db *sqlx.DB) (CreateParams, func(*testing.T)) {
				email := uniqueEmail(t, db)
				if _, err := Create(context.Background(), db, CreateParams{Email: email, Name: "First", Password: "pass"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
				return CreateParams{Email: email, Name: "Second", Password: "pass"}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, check := tt.arrange(t, db)
			got, err := Create(context.Background(), db, params)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil {
				if got.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if got.Email != params.Email {
					t.Fatalf("email: got %q, want %q", got.Email, params.Email)
				}
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns existing user",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.ID, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "00000000-0000-0000-0000-000000000000", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, check := tt.arrange(t, db)
			got, err := Get(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.ID != id {
				t.Fatalf("id: got %q, want %q", got.ID, id)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) (string, func(*testing.T))
		wantErr error
	}{
		{
			name: "returns user by email",
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				u := seedUser(t, db)
				return u.Email, func(t *testing.T) {}
			},
		},
		{
			name:    "not found",
			wantErr: ErrNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) (string, func(*testing.T)) {
				return "nobody@does-not-exist.local", nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, check := tt.arrange(t, db)
			got, err := GetByEmail(context.Background(), db, email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserByEmail() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got.Email != email {
				t.Fatalf("email: got %q, want %q", got.Email, email)
			}
			if check != nil {
				check(t)
			}
		})
	}
}

func TestArchiveUser(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	tests := []struct {
		name    string
		arrange func(*testing.T, *sqlx.DB) string
		wantErr error
	}{
		{
			name: "archives active user",
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return seedUser(t, db).ID
			},
		},
		{
			name:    "not found",
			wantErr: ErrNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				return "00000000-0000-0000-0000-000000000000"
			},
		},
		{
			name:    "already archived",
			wantErr: ErrNotFound,
			arrange: func(t *testing.T, db *sqlx.DB) string {
				u := seedUser(t, db)
				if err := Archive(context.Background(), db, u.ID); err != nil {
					t.Fatalf("first archive: %v", err)
				}
				return u.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.arrange(t, db)
			err := Archive(context.Background(), db, id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ArchiveUser() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// --- reset token integration tests ---

func TestCreateAndValidateResetToken(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, err := CreateResetToken(ctx, db, userID)
	if err != nil {
		t.Fatalf("CreateResetToken error = %v", err)
	}
	if rawToken == "" {
		t.Fatal("CreateResetToken returned empty token")
	}

	gotUserID, err := ValidateResetToken(ctx, db, rawToken)
	if err != nil {
		t.Fatalf("ValidateResetToken error = %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("ValidateResetToken userID = %q, want %q", gotUserID, userID)
	}
}

func TestValidateResetToken_NotFound(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	_, err := ValidateResetToken(context.Background(), db, "nonexistent_token")
	if !errors.Is(err, ErrResetTokenNotFound) {
		t.Fatalf("error = %v, want ErrResetTokenNotFound", err)
	}
}

func TestValidateResetToken_Used(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateResetToken(ctx, db, userID)
	hash := sessions.HashToken(rawToken)
	db.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`, hash)

	_, err := ValidateResetToken(ctx, db, rawToken)
	if !errors.Is(err, ErrResetTokenUsed) {
		t.Fatalf("error = %v, want ErrResetTokenUsed", err)
	}
}

func TestValidateResetToken_Expired(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateResetToken(ctx, db, userID)
	hash := sessions.HashToken(rawToken)
	db.ExecContext(ctx, `UPDATE password_reset_tokens SET expires_at = NOW() - INTERVAL '1 hour' WHERE token_hash = $1`, hash)

	_, err := ValidateResetToken(ctx, db, rawToken)
	if !errors.Is(err, ErrResetTokenExpired) {
		t.Fatalf("error = %v, want ErrResetTokenExpired", err)
	}
}

func TestResetPassword_EndToEnd(t *testing.T) {
	db := testpg.Open(t)
	testpg.EnsureMigrated(t, db)

	userID := testpg.SeedUser(t, db)
	ctx := context.Background()

	rawToken, _ := CreateResetToken(ctx, db, userID)

	err := ResetPassword(ctx, db, rawToken, "newpassword123")
	if err != nil {
		t.Fatalf("ResetPassword error = %v", err)
	}

	// Token should now be used
	_, err = ValidateResetToken(ctx, db, rawToken)
	if !errors.Is(err, ErrResetTokenUsed) {
		t.Fatalf("token should be used after reset, got error = %v", err)
	}
}

// --- helpers ---

func uniqueEmail(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	suffix := testpg.UniqueSuffix(t, db)
	return fmt.Sprintf("user-%s@test.local", suffix)
}

func seedUser(t *testing.T, db *sqlx.DB) User {
	t.Helper()
	u, err := Create(context.Background(), db, CreateParams{
		Email:    uniqueEmail(t, db),
		Name:     "Test User",
		Password: "testpass123",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), `DELETE FROM app_users WHERE id = $1`, u.ID)
	})
	return u
}
