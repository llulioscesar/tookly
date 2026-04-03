// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/pgutil"
	"github.com/start-codex/tookly/internal/sessions"
)

const VerifyTokenTTL = 24 * time.Hour

var (
	ErrVerifyTokenNotFound = errors.New("verification token not found")
	ErrVerifyTokenExpired  = errors.New("verification token expired")
	ErrVerifyTokenUsed     = errors.New("verification token already used")
)

func CreateVerifyToken(ctx context.Context, db *sqlx.DB, userID string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if userID == "" {
		return "", errors.New("userID is required")
	}
	rawToken, err := sessions.GenerateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	tokenHash := sessions.HashToken(rawToken)
	expiresAt := time.Now().Add(VerifyTokenTTL)
	if err := createVerifyToken(ctx, db, userID, tokenHash, expiresAt); err != nil {
		return "", err
	}
	return rawToken, nil
}

func VerifyEmail(ctx context.Context, db *sqlx.DB, rawToken string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if rawToken == "" {
		return errors.New("token is required")
	}
	tokenHash := sessions.HashToken(rawToken)
	token, err := getVerifyTokenByHash(ctx, db, tokenHash)
	if err != nil {
		return err
	}
	if token.UsedAt != nil {
		return ErrVerifyTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return ErrVerifyTokenExpired
	}

	return pgutil.WithTx(ctx, db, nil, "begin tx", "commit", func(tx *sqlx.Tx) error {
		if err := setEmailVerifiedTx(ctx, tx, token.UserID); err != nil {
			return fmt.Errorf("set verified: %w", err)
		}
		return markVerifyTokenUsedTx(ctx, tx, tokenHash)
	})
}

func IsVerificationRequired(ctx context.Context, db *sqlx.DB) (bool, error) {
	var val string
	err := db.GetContext(ctx, &val,
		`SELECT value FROM instance_config WHERE key = 'email_verification_required'`)
	if err != nil {
		// Key missing or any DB error — treat as not required
		return false, nil
	}
	if val != "true" && val != "false" {
		return false, fmt.Errorf("invalid email_verification_required value: %q", val)
	}
	return val == "true", nil
}

// SendVerificationEmail creates a token and sends the verification email.
func SendVerificationEmail(ctx context.Context, db *sqlx.DB, userID, recipientEmail, baseURL string) error {
	rawToken, err := CreateVerifyToken(ctx, db, userID)
	if err != nil {
		return fmt.Errorf("create verification token: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, rawToken)

	body, err := email.RenderTemplate("email_verification", struct{ VerifyURL string }{verifyURL})
	if err != nil {
		return fmt.Errorf("render verification email: %w", err)
	}

	smtpConfig, err := loadSMTPConfig(ctx, db)
	if err != nil {
		return fmt.Errorf("smtp not configured: %w", err)
	}
	if err := email.Send(smtpConfig, email.Message{
		To:      recipientEmail,
		Subject: "Verify your Tookly email",
		Body:    body,
	}); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}
	return nil
}
