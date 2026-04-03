// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package issuetypes

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound  = errors.New("issue type not found")
	ErrDuplicate = errors.New("issue type name already exists in project")
)

type Type struct {
	ID         string     `db:"id"          json:"id"`
	ProjectID  string     `db:"project_id"  json:"project_id"`
	Name       string     `db:"name"        json:"name"`
	Icon       string     `db:"icon"        json:"icon"`
	Level      int        `db:"level"       json:"level"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type CreateParams struct {
	ProjectID string
	Name      string
	Icon      string
	Level     int
}

func (params CreateParams) Validate() error {
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.Name == "" {
		return errors.New("name is required")
	}
	if params.Level < 0 {
		return errors.New("level must be >= 0")
	}
	return nil
}

func Create(ctx context.Context, db *sqlx.DB, params CreateParams) (Type, error) {
	if db == nil {
		return Type{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Type{}, err
	}
	return createIssueType(ctx, db, params)
}

func List(ctx context.Context, db *sqlx.DB, projectID string) ([]Type, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listIssueTypes(ctx, db, projectID)
}

func Archive(ctx context.Context, db *sqlx.DB, projectID, issueTypeID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if issueTypeID == "" {
		return errors.New("issue_type_id is required")
	}
	return archiveIssueType(ctx, db, projectID, issueTypeID)
}
