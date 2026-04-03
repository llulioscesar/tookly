// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

const sessionCols = `s.id, s.user_id, s.created_at, s.expires_at, s.last_used_at`

func createSession(ctx context.Context, db *sqlx.DB, userID string, ttl time.Duration) (CreateResult, error) {
	rawToken, err := GenerateToken()
	if err != nil {
		return CreateResult{}, fmt.Errorf("generate token: %w", err)
	}

	hashedToken := HashToken(rawToken)
	now := time.Now()
	expiresAt := now.Add(ttl)

	var session Session
	err = db.QueryRowxContext(ctx,
		`INSERT INTO sessions (id, user_id, created_at, expires_at, last_used_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, created_at, expires_at, last_used_at`,
		hashedToken, userID, now, expiresAt, now,
	).StructScan(&session)
	if err != nil {
		return CreateResult{}, fmt.Errorf("insert session: %w", err)
	}
	return CreateResult{Session: session, RawToken: rawToken}, nil
}

func createSessionTx(ctx context.Context, tx *sqlx.Tx, userID string, ttl time.Duration) (CreateResult, error) {
	rawToken, err := GenerateToken()
	if err != nil {
		return CreateResult{}, fmt.Errorf("generate token: %w", err)
	}

	hashedToken := HashToken(rawToken)
	now := time.Now()
	expiresAt := now.Add(ttl)

	var session Session
	err = tx.QueryRowxContext(ctx,
		`INSERT INTO sessions (id, user_id, created_at, expires_at, last_used_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, created_at, expires_at, last_used_at`,
		hashedToken, userID, now, expiresAt, now,
	).StructScan(&session)
	if err != nil {
		return CreateResult{}, fmt.Errorf("insert session: %w", err)
	}
	return CreateResult{Session: session, RawToken: rawToken}, nil
}

func validateSession(ctx context.Context, db *sqlx.DB, rawToken string) (Session, error) {
	hashedToken := HashToken(rawToken)

	var row struct {
		Session
		UserArchived bool `db:"user_archived"`
	}
	err := db.GetContext(ctx, &row,
		`SELECT `+sessionCols+`, (u.archived_at IS NOT NULL) AS user_archived
		 FROM sessions s
		 LEFT JOIN app_users u ON u.id = s.user_id
		 WHERE s.id = $1`,
		hashedToken,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("get session: %w", err)
	}

	if row.UserArchived {
		return Session{}, ErrUserArchived
	}

	if time.Now().After(row.Session.ExpiresAt) {
		return Session{}, ErrSessionExpired
	}

	return row.Session, nil
}

func deleteByUserID(ctx context.Context, db *sqlx.DB, userID, exceptTokenHash string) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = $1 AND id != $2`,
		userID, exceptTokenHash,
	)
	if err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}

func deleteSession(ctx context.Context, db *sqlx.DB, rawToken string) error {
	hashedToken := HashToken(rawToken)
	_, err := db.ExecContext(ctx,
		`DELETE FROM sessions WHERE id = $1`,
		hashedToken,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
