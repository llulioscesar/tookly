// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/pgutil"
)

// --- user store ---

const userCols = `id, email, name, is_instance_admin, email_verified_at, created_at, updated_at, archived_at, password_hash`

func createUser(ctx context.Context, db *sqlx.DB, params CreateParams) (User, error) {
	hash, err := hashPassword(params.Password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	var user User
	err = db.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING `+userCols,
		params.Email, params.Name, hash,
	).StructScan(&user)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func createInstanceAdminTx(ctx context.Context, tx *sqlx.Tx, params CreateParams) (User, error) {
	hash, err := hashPassword(params.Password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	var user User
	err = tx.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name, password_hash, is_instance_admin, email_verified_at)
		 VALUES ($1, $2, $3, true, NOW())
		 RETURNING `+userCols,
		params.Email, params.Name, hash,
	).StructScan(&user)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert instance admin: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func getUser(ctx context.Context, db *sqlx.DB, id string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func getUserTx(ctx context.Context, tx *sqlx.Tx, id string) (User, error) {
	var user User
	err := tx.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE id = $1`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user tx: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func getUserByEmailTx(ctx context.Context, tx *sqlx.Tx, email string) (User, error) {
	var user User
	err := tx.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by email tx: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func getUserByEmail(ctx context.Context, db *sqlx.DB, email string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func authenticateUser(ctx context.Context, db *sqlx.DB, email, password string) (User, error) {
	var user User
	err := db.GetContext(ctx, &user,
		`SELECT `+userCols+` FROM app_users WHERE email = $1`,
		email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("get user for auth: %w", err)
	}
	// OIDC-only users have no password hash — treat as invalid credentials
	if user.PasswordHash == "" {
		return User{}, ErrInvalidCredentials
	}
	ok, err := verifyPassword(user.PasswordHash, password)
	if err != nil {
		return User{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func createOIDCUser(ctx context.Context, db *sqlx.DB, params CreateOIDCUserParams) (User, error) {
	var user User
	err := db.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name, password_hash, email_verified_at)
		 VALUES ($1, $2, '', NOW())
		 RETURNING `+userCols,
		params.Email, params.Name,
	).StructScan(&user)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert oidc user: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func createOIDCUserTx(ctx context.Context, tx *sqlx.Tx, params CreateOIDCUserParams) (User, error) {
	var user User
	err := tx.QueryRowxContext(ctx,
		`INSERT INTO app_users (email, name, password_hash, email_verified_at)
		 VALUES ($1, $2, '', NOW())
		 RETURNING `+userCols,
		params.Email, params.Name,
	).StructScan(&user)
	if err != nil {
		if pgutil.IsUniqueViolation(err) {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("insert oidc user tx: %w", err)
	}
	user.fillDerived()
	user.PasswordHash = ""
	return user, nil
}

func getPasswordHash(ctx context.Context, db *sqlx.DB, userID string) (string, error) {
	var hash string
	err := db.GetContext(ctx, &hash,
		`SELECT password_hash FROM app_users WHERE id = $1 AND archived_at IS NULL`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get password hash: %w", err)
	}
	return hash, nil
}

func updatePassword(ctx context.Context, db *sqlx.DB, userID, newHash string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE app_users SET password_hash = $2, updated_at = NOW() WHERE id = $1 AND archived_at IS NULL`,
		userID, newHash,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update password rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func updatePasswordTx(ctx context.Context, tx *sqlx.Tx, userID, newHash string) error {
	res, err := tx.ExecContext(ctx,
		`UPDATE app_users SET password_hash = $2, updated_at = NOW() WHERE id = $1 AND archived_at IS NULL`,
		userID, newHash,
	)
	if err != nil {
		return fmt.Errorf("update password tx: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update password tx rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func archiveUser(ctx context.Context, db *sqlx.DB, id string) error {
	res, err := db.ExecContext(ctx,
		`UPDATE app_users
		 SET archived_at = NOW()
		 WHERE id = $1 AND archived_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("archive user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive user rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- verify token store ---

type verificationToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func createVerifyToken(ctx context.Context, db *sqlx.DB, userID, tokenHash string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert verification token: %w", err)
	}
	return nil
}

func getVerifyTokenByHash(ctx context.Context, db *sqlx.DB, tokenHash string) (verificationToken, error) {
	var token verificationToken
	err := db.GetContext(ctx, &token,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM email_verification_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return verificationToken{}, ErrVerifyTokenNotFound
		}
		return verificationToken{}, fmt.Errorf("get verification token: %w", err)
	}
	return token, nil
}

func markVerifyTokenUsedTx(ctx context.Context, tx *sqlx.Tx, tokenHash string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mark verification token used: %w", err)
	}
	return nil
}

func setEmailVerifiedTx(ctx context.Context, tx *sqlx.Tx, userID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE app_users SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1 AND email_verified_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}
	return nil
}

// --- reset token store ---

func createResetToken(ctx context.Context, db *sqlx.DB, userID, tokenHash string, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert reset token: %w", err)
	}
	return nil
}

func getResetTokenByHash(ctx context.Context, db *sqlx.DB, tokenHash string) (ResetToken, error) {
	var token ResetToken
	err := db.GetContext(ctx, &token,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ResetToken{}, ErrResetTokenNotFound
		}
		return ResetToken{}, fmt.Errorf("get reset token: %w", err)
	}
	return token, nil
}

func markResetTokenUsedTx(ctx context.Context, tx *sqlx.Tx, tokenHash string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("mark token used tx: %w", err)
	}
	return nil
}

// --- instance config helpers (avoids import cycle with internal/instance) ---

func getInstanceConfig(ctx context.Context, db *sqlx.DB, key string) (string, bool) {
	var val string
	err := db.GetContext(ctx, &val, `SELECT value FROM instance_config WHERE key = $1`, key)
	return val, err == nil
}

func loadSMTPConfig(ctx context.Context, db *sqlx.DB) (*email.SMTPConfig, error) {
	host, ok := getInstanceConfig(ctx, db, "smtp_host")
	if !ok || host == "" {
		return nil, email.ErrSMTPNotConfigured
	}
	portStr, _ := getInstanceConfig(ctx, db, "smtp_port")
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 587
	}
	from, _ := getInstanceConfig(ctx, db, "smtp_from")
	username, _ := getInstanceConfig(ctx, db, "smtp_username")
	password, _ := getInstanceConfig(ctx, db, "smtp_password")
	return &email.SMTPConfig{
		Host:     host,
		Port:     port,
		From:     from,
		Username: username,
		Password: password,
	}, nil
}

// resolveBaseURL returns the effective base URL for building absolute links.
// Priority: configured base_url → Origin header → X-Forwarded-Proto + Host → http + Host.
func resolveBaseURL(ctx context.Context, db *sqlx.DB, r *http.Request) string {
	if baseURL, ok := getInstanceConfig(ctx, db, "base_url"); ok && baseURL != "" {
		return baseURL
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	return fmt.Sprintf("%s://%s", proto, r.Host)
}
