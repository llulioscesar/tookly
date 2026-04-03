// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package projects

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound       = errors.New("project not found")
	ErrDuplicateKey   = errors.New("project key already exists in workspace")
	ErrMemberNotFound = errors.New("member not found")
)

var validRoles = map[string]bool{"admin": true, "member": true, "viewer": true}

var reKey = regexp.MustCompile(`^[A-Z]{2,10}$`)

type Project struct {
	ID          string     `db:"id"           json:"id"`
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	Name        string     `db:"name"         json:"name"`
	Key         string     `db:"key"          json:"key"`
	Description string     `db:"description"  json:"description"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
	ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
}

var validTemplates = map[string]bool{"kanban": true, "scrum": true}

type CreateParams struct {
	WorkspaceID string
	Name        string
	Key         string
	Description string
	Template    string
	Locale      string
}

func (params CreateParams) Validate() error {
	if params.WorkspaceID == "" {
		return errors.New("workspace_id is required")
	}
	if params.Name == "" {
		return errors.New("name is required")
	}
	if !reKey.MatchString(params.Key) {
		return errors.New("key must be 2-10 uppercase letters (A-Z)")
	}
	if params.Template != "" && !validTemplates[params.Template] {
		return errors.New("template must be 'kanban' or 'scrum'")
	}
	return nil
}

func Create(ctx context.Context, db *sqlx.DB, params CreateParams) (Project, error) {
	if db == nil {
		return Project{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Project{}, err
	}
	return createProject(ctx, db, params)
}

func Get(ctx context.Context, db *sqlx.DB, id string) (Project, error) {
	if db == nil {
		return Project{}, errors.New("db is required")
	}
	if id == "" {
		return Project{}, errors.New("id is required")
	}
	return getProject(ctx, db, id)
}

func List(ctx context.Context, db *sqlx.DB, workspaceID string) ([]Project, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	return listProjects(ctx, db, workspaceID)
}

type Member struct {
	ProjectID  string     `db:"project_id"  json:"project_id"`
	UserID     string     `db:"user_id"     json:"user_id"`
	Role       string     `db:"role"        json:"role"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updated_at"`
	ArchivedAt *time.Time `db:"archived_at" json:"archived_at,omitempty"`
}

type AddMemberParams struct {
	ProjectID string
	UserID    string
	Role      string
}

func (params AddMemberParams) Validate() error {
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validRoles[params.Role] {
		return errors.New("role must be 'admin', 'member' or 'viewer'")
	}
	return nil
}

type UpdateMemberRoleParams struct {
	ProjectID string
	UserID    string
	Role      string
}

func (params UpdateMemberRoleParams) Validate() error {
	if params.ProjectID == "" {
		return errors.New("project_id is required")
	}
	if params.UserID == "" {
		return errors.New("user_id is required")
	}
	if !validRoles[params.Role] {
		return errors.New("role must be 'admin', 'member' or 'viewer'")
	}
	return nil
}

func AddMember(ctx context.Context, db *sqlx.DB, params AddMemberParams) (Member, error) {
	if db == nil {
		return Member{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Member{}, err
	}
	return addMember(ctx, db, params)
}

func RemoveMember(ctx context.Context, db *sqlx.DB, projectID, userID string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if projectID == "" {
		return errors.New("project_id is required")
	}
	if userID == "" {
		return errors.New("user_id is required")
	}
	return removeMember(ctx, db, projectID, userID)
}

func ListMembers(ctx context.Context, db *sqlx.DB, projectID string) ([]Member, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if projectID == "" {
		return nil, errors.New("project_id is required")
	}
	return listMembers(ctx, db, projectID)
}

func UpdateMemberRole(ctx context.Context, db *sqlx.DB, params UpdateMemberRoleParams) (Member, error) {
	if db == nil {
		return Member{}, errors.New("db is required")
	}
	if err := params.Validate(); err != nil {
		return Member{}, err
	}
	return updateMemberRole(ctx, db, params)
}

func Archive(ctx context.Context, db *sqlx.DB, id string) error {
	if db == nil {
		return errors.New("db is required")
	}
	if id == "" {
		return errors.New("id is required")
	}
	return archiveProject(ctx, db, id)
}
