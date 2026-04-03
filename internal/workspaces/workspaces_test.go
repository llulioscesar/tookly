// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package workspaces

import (
	"context"
	"testing"
)

func TestCreateWorkspaceParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateParams
		wantErr bool
	}{
		{
			name:    "valid",
			params:  CreateParams{Name: "Acme Corp", Slug: "acme-corp", OwnerID: "user-1"},
			wantErr: false,
		},
		{
			name:    "slug exactly 2 chars",
			params:  CreateParams{Name: "AB", Slug: "ab", OwnerID: "user-1"},
			wantErr: false,
		},
		{
			name:    "slug with digits",
			params:  CreateParams{Name: "Team 42", Slug: "team42", OwnerID: "user-1"},
			wantErr: false,
		},
		{
			name:    "missing name",
			params:  CreateParams{Name: "", Slug: "acme"},
			wantErr: true,
		},
		{
			name:    "slug too short",
			params:  CreateParams{Name: "A", Slug: "a"},
			wantErr: true,
		},
		{
			name:    "slug with uppercase",
			params:  CreateParams{Name: "Acme", Slug: "Acme"},
			wantErr: true,
		},
		{
			name:    "slug starts with hyphen",
			params:  CreateParams{Name: "Acme", Slug: "-acme"},
			wantErr: true,
		},
		{
			name:    "slug with spaces",
			params:  CreateParams{Name: "Acme", Slug: "acme corp"},
			wantErr: true,
		},
		{
			name:    "empty slug",
			params:  CreateParams{Name: "Acme", Slug: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateWorkspace_NilDB(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateParams{Name: "Acme", Slug: "acme"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateWorkspace() error = %v, want %q", err, "db is required")
	}
}

func TestGetWorkspace_NilDB(t *testing.T) {
	_, err := Get(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetWorkspace() error = %v, want %q", err, "db is required")
	}
}

func TestGetWorkspace_EmptyID(t *testing.T) {
	_, err := Get(context.Background(), nil, "")
	if err == nil {
		t.Fatal("GetWorkspace() with empty id should return error")
	}
}

func TestGetWorkspaceBySlug_NilDB(t *testing.T) {
	_, err := GetBySlug(context.Background(), nil, "acme")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetWorkspaceBySlug() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveWorkspace_NilDB(t *testing.T) {
	err := Archive(context.Background(), nil, "some-id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveWorkspace() error = %v, want %q", err, "db is required")
	}
}
