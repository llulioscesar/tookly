// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package auth

import (
	"context"
	"testing"
)

func TestCreateUserParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateParams
		wantErr bool
	}{
		{name: "valid", params: CreateParams{Email: "alice@example.com", Name: "Alice", Password: "secretpass"}, wantErr: false},
		{name: "missing name", params: CreateParams{Email: "alice@example.com", Name: "", Password: "secretpass"}, wantErr: true},
		{name: "missing email", params: CreateParams{Email: "", Name: "Alice", Password: "secretpass"}, wantErr: true},
		{name: "email without @", params: CreateParams{Email: "notanemail", Name: "Alice", Password: "secretpass"}, wantErr: true},
		{name: "missing password", params: CreateParams{Email: "alice@example.com", Name: "Alice", Password: ""}, wantErr: true},
		{name: "password too short", params: CreateParams{Email: "alice@example.com", Name: "Alice", Password: "short"}, wantErr: true},
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

func TestCreateUser_NilDB(t *testing.T) {
	_, err := Create(context.Background(), nil, CreateParams{Email: "a@b.com", Name: "A", Password: "p"})
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("CreateUser() error = %v, want %q", err, "db is required")
	}
}

func TestGetUser_NilDB(t *testing.T) {
	_, err := Get(context.Background(), nil, "id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetUser() error = %v, want %q", err, "db is required")
	}
}

func TestGetUserByEmail_NilDB(t *testing.T) {
	_, err := GetByEmail(context.Background(), nil, "a@b.com")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("GetUserByEmail() error = %v, want %q", err, "db is required")
	}
}

func TestArchiveUser_NilDB(t *testing.T) {
	err := Archive(context.Background(), nil, "id")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("ArchiveUser() error = %v, want %q", err, "db is required")
	}
}

func TestAuthenticateUser_NilDB(t *testing.T) {
	_, err := Authenticate(context.Background(), nil, "a@b.com", "pass")
	if err == nil || err.Error() != "db is required" {
		t.Fatalf("AuthenticateUser() error = %v, want %q", err, "db is required")
	}
}
