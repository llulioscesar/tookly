// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package issues

import (
	"context"
	"testing"
	"time"
)

func TestMoveIssueParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  MoveParams
		wantErr bool
	}{
		{
			name:    "valid params",
			params:  MoveParams{ProjectID: "proj-1", IssueID: "issue-1", TargetPosition: 0},
			wantErr: false,
		},
		{
			name:    "missing project_id",
			params:  MoveParams{ProjectID: "", IssueID: "issue-1", TargetPosition: 0},
			wantErr: true,
		},
		{
			name:    "missing issue_id",
			params:  MoveParams{ProjectID: "proj-1", IssueID: "", TargetPosition: 0},
			wantErr: true,
		},
		{
			name:    "negative target_position",
			params:  MoveParams{ProjectID: "proj-1", IssueID: "issue-1", TargetPosition: -1},
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

func TestMoveIssue_NilDB(t *testing.T) {
	err := Move(context.Background(), nil, MoveParams{
		ProjectID:      "proj-1",
		IssueID:        "issue-1",
		TargetPosition: 0,
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("Move() error = %v, want %q", err, "db is required")
	}
}

func TestCreateIssueParams_Validate(t *testing.T) {
	due := time.Now()
	valid := CreateParams{
		ProjectID:   "p",
		IssueTypeID: "t",
		StatusID:    "s",
		Title:       "Fix bug",
		ReporterID:  "r",
		Priority:    "high",
	}

	tests := []struct {
		name    string
		params  CreateParams
		wantErr bool
	}{
		{name: "valid", params: valid, wantErr: false},
		{name: "priority defaults to medium", params: func() CreateParams { c := valid; c.Priority = ""; return c }(), wantErr: false},
		{name: "valid with due date", params: func() CreateParams { c := valid; c.DueDate = &due; return c }(), wantErr: false},
		{name: "missing project_id", params: func() CreateParams { c := valid; c.ProjectID = ""; return c }(), wantErr: true},
		{name: "missing issue_type_id", params: func() CreateParams { c := valid; c.IssueTypeID = ""; return c }(), wantErr: true},
		{name: "missing status_id", params: func() CreateParams { c := valid; c.StatusID = ""; return c }(), wantErr: true},
		{name: "missing title", params: func() CreateParams { c := valid; c.Title = ""; return c }(), wantErr: true},
		{name: "missing reporter_id", params: func() CreateParams { c := valid; c.ReporterID = ""; return c }(), wantErr: true},
		{name: "invalid priority", params: func() CreateParams { c := valid; c.Priority = "urgent"; return c }(), wantErr: true},
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

func TestUpdateIssueParams_Validate(t *testing.T) {
	valid := UpdateParams{
		IssueID:   "i",
		ProjectID: "p",
		Title:     "Fix bug",
		Priority:  "low",
	}

	tests := []struct {
		name    string
		params  UpdateParams
		wantErr bool
	}{
		{name: "valid", params: valid, wantErr: false},
		{name: "missing issue_id", params: func() UpdateParams { c := valid; c.IssueID = ""; return c }(), wantErr: true},
		{name: "missing project_id", params: func() UpdateParams { c := valid; c.ProjectID = ""; return c }(), wantErr: true},
		{name: "missing title", params: func() UpdateParams { c := valid; c.Title = ""; return c }(), wantErr: true},
		{name: "invalid priority", params: func() UpdateParams { c := valid; c.Priority = "asap"; return c }(), wantErr: true},
		{name: "empty priority invalid", params: func() UpdateParams { c := valid; c.Priority = ""; return c }(), wantErr: true},
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

func TestCreateIssue_NilDB(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateParams{
		ProjectID: "p", IssueTypeID: "t", StatusID: "s", Title: "T", ReporterID: "r", Priority: "medium",
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("Create() error = %v, want %q", err, "db is required")
	}
}

func TestGetIssue_NilDB(t *testing.T) {
	_, err := Get(context.Background(), nil, "p", "i")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("Get() error = %v, want %q", err, "db is required")
	}
}

func TestListIssues_NilDB(t *testing.T) {
	_, err := List(context.Background(), nil, ListParams{ProjectID: "p"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("List() error = %v, want %q", err, "db is required")
	}
}

func TestUpdateIssue_NilDB(t *testing.T) {
	_, err := Update(context.Background(), nil, UpdateParams{
		IssueID: "i", ProjectID: "p", Title: "T", Priority: "medium",
	})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("Update() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveIssue_NilDB(t *testing.T) {
	err := Archive(context.Background(), nil, "p", "i")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("Archive() error = %v, want %q", err, "db is required")
	}
}
