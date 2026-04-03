// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package instance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/auth"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/sessions"
)

var (
	ErrConfigNotFound     = errors.New("config key not found")
	ErrAlreadyInitialized = errors.New("instance already initialized")
)

type BootstrapParams struct {
	Email    string
	Name     string
	Password string
}

func (p BootstrapParams) Validate() error {
	return auth.CreateParams{
		Email:    p.Email,
		Name:     p.Name,
		Password: p.Password,
	}.Validate()
}

type BootstrapResult struct {
	User     auth.User
	RawToken string
}

func Bootstrap(ctx context.Context, db *sqlx.DB, params BootstrapParams) (BootstrapResult, error) {
	if db == nil {
		return BootstrapResult{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return BootstrapResult{}, err
	}

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock the initialized row to prevent concurrent bootstrap
	var val string
	err = tx.GetContext(ctx, &val,
		`SELECT value FROM instance_config WHERE key = 'initialized' FOR UPDATE`)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("lock initialized: %w", err)
	}
	if val == "true" {
		return BootstrapResult{}, ErrAlreadyInitialized
	}
	if val != "false" {
		return BootstrapResult{}, fmt.Errorf("invalid initialized value: %q — cannot bootstrap", val)
	}

	// Create the instance admin user
	user, err := auth.CreateInstanceAdminTx(ctx, tx, auth.CreateParams{
		Email:    params.Email,
		Name:     params.Name,
		Password: params.Password,
	})
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("create admin: %w", err)
	}

	// Create session for the new admin
	sessionResult, err := sessions.CreateTx(ctx, tx, user.ID)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("create session: %w", err)
	}

	// Mark instance as initialized
	_, err = tx.ExecContext(ctx,
		`UPDATE instance_config SET value = 'true', updated_at = NOW() WHERE key = 'initialized'`)
	if err != nil {
		return BootstrapResult{}, fmt.Errorf("set initialized: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return BootstrapResult{}, fmt.Errorf("commit: %w", err)
	}

	return BootstrapResult{User: user, RawToken: sessionResult.RawToken}, nil
}

func GetConfig(ctx context.Context, db *sqlx.DB, key string) (string, error) {
	if db == nil {
		return "", errors.New("db is required")
	}
	if key == "" {
		return "", errors.New("key is required")
	}
	return getConfig(ctx, db, key)
}

func SetConfig(ctx context.Context, db *sqlx.DB, key, value string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if key == "" {
		return errors.New("key is required")
	}
	return setConfig(ctx, db, key, value)
}

func IsInitialized(ctx context.Context, db *sqlx.DB) (bool, error) {
	val, err := GetConfig(ctx, db, "initialized")
	if err != nil {
		return false, err
	}
	if val != "true" && val != "false" {
		return false, fmt.Errorf("invalid initialized value: %q", val)
	}
	return val == "true", nil
}

func LoadSMTPConfig(ctx context.Context, db *sqlx.DB) (*email.SMTPConfig, error) {
	host, err := GetConfig(ctx, db, "smtp_host")
	if err != nil {
		if errors.Is(err, ErrConfigNotFound) {
			return nil, email.ErrSMTPNotConfigured
		}
		return nil, err
	}
	if host == "" {
		return nil, email.ErrSMTPNotConfigured
	}

	portStr, _ := GetConfig(ctx, db, "smtp_port")
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 587
	}

	from, _ := GetConfig(ctx, db, "smtp_from")
	username, _ := GetConfig(ctx, db, "smtp_username")
	password, _ := GetConfig(ctx, db, "smtp_password")

	return &email.SMTPConfig{
		Host:     host,
		Port:     port,
		From:     from,
		Username: username,
		Password: password,
	}, nil
}

// ResolveBaseURL returns the effective base URL for building absolute links.
// Priority: configured base_url → Origin header → X-Forwarded-Proto + Host → http + Host.
func ResolveBaseURL(ctx context.Context, db *sqlx.DB, r *http.Request) string {
	baseURL, _ := GetConfig(ctx, db, "base_url")
	if baseURL != "" {
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

func SaveSMTPConfig(ctx context.Context, db *sqlx.DB, config email.SMTPConfig) error {
	keys := map[string]string{
		"smtp_host":     config.Host,
		"smtp_port":     strconv.Itoa(config.Port),
		"smtp_from":     config.From,
		"smtp_username": config.Username,
		"smtp_password": config.Password,
	}
	for k, v := range keys {
		if err := SetConfig(ctx, db, k, v); err != nil {
			return err
		}
	}
	return nil
}
