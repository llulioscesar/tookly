// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package invitations

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/auth"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/instance"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/workspaces"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	mux.HandleFunc("POST /workspaces/{workspaceID}/invitations", handleCreate(db))
	mux.HandleFunc("GET /workspaces/{workspaceID}/invitations", handleListPending(db))
	mux.HandleFunc("DELETE /invitations/{invitationID}", handleRevoke(db))
	mux.HandleFunc("POST /invitations/{invitationID}/resend", handleResend(db))
	mux.HandleFunc("GET /invitations/accept", handleGetAccept(db))
	mux.HandleFunc("POST /invitations/accept", handleAccept(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, authz.ErrUnauthenticated):
		respond.Error(w, http.StatusUnauthorized, "authentication required")
	case errors.Is(err, authz.ErrForbidden):
		respond.Error(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, authz.ErrWorkspaceNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, "invitation not found")
	case errors.Is(err, ErrDuplicate):
		respond.Error(w, http.StatusConflict, "pending invitation already exists for this email")
	case errors.Is(err, ErrAlreadyMember):
		respond.Error(w, http.StatusConflict, "user is already a workspace member")
	case errors.Is(err, ErrExpired), errors.Is(err, ErrRevoked), errors.Is(err, ErrUsed):
		respond.Error(w, http.StatusBadRequest, "invalid_or_expired_invitation")
	case errors.Is(err, auth.ErrDuplicateEmail):
		respond.Error(w, http.StatusConflict, "email already exists")
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		userID, _ := authz.UserIDFromContext(r.Context())

		var body struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		params := CreateParams{
			WorkspaceID: wsID,
			Email:       body.Email,
			Role:        body.Role,
			InvitedBy:   userID,
		}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		rawToken, inv, err := Create(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}

		// Send invitation email
		sendInvitationEmail(r, db, rawToken, inv, userID)

		respond.JSON(w, http.StatusCreated, inv)
	}
}

func handleListPending(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsID := r.PathValue("workspaceID")
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, wsID); err != nil {
			fail(w, err)
			return
		}
		invs, err := ListPending(r.Context(), db, wsID)
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, invs)
	}
}

func handleRevoke(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invID := r.PathValue("invitationID")
		inv, err := GetByID(r.Context(), db, invID)
		if err != nil {
			fail(w, err)
			return
		}
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, inv.WorkspaceID); err != nil {
			fail(w, err)
			return
		}
		if err := Revoke(r.Context(), db, invID); err != nil {
			fail(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleResend(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invID := r.PathValue("invitationID")
		inv, err := GetByID(r.Context(), db, invID)
		if err != nil {
			fail(w, err)
			return
		}
		if err := authz.RequireWorkspaceAdmin(r.Context(), db, inv.WorkspaceID); err != nil {
			fail(w, err)
			return
		}
		userID, _ := authz.UserIDFromContext(r.Context())

		rawToken, err := Resend(r.Context(), db, invID)
		if err != nil {
			fail(w, err)
			return
		}

		// Re-send email with new token
		sendInvitationEmail(r, db, rawToken, inv, userID)

		respond.JSON(w, http.StatusOK, map[string]string{"status": "resent"})
	}
}

func handleGetAccept(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			respond.Error(w, http.StatusBadRequest, "token is required")
			return
		}
		inv, err := Get(r.Context(), db, token)
		if err != nil {
			fail(w, err)
			return
		}
		// Get workspace name and inviter name for display
		ws, _ := workspaces.Get(r.Context(), db, inv.WorkspaceID)
		inviter, _ := auth.Get(r.Context(), db, inv.InvitedBy)
		wsName := ""
		if ws.ID != "" {
			wsName = ws.Name
		}
		inviterName := ""
		if inviter.ID != "" {
			inviterName = inviter.Name
		}
		respond.JSON(w, http.StatusOK, map[string]string{
			"email":          inv.Email,
			"role":           inv.Role,
			"workspace_name": wsName,
			"inviter_name":   inviterName,
		})
	}
}

func handleAccept(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token    string `json:"token"`
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Token == "" {
			respond.Error(w, http.StatusBadRequest, "token is required")
			return
		}

		// Look up invitation before accept (token becomes unusable after)
		inv, invErr := Get(r.Context(), db, body.Token)
		if invErr != nil {
			fail(w, invErr)
			return
		}

		var userID string

		// Check if authenticated (existing user accepting)
		authedUserID, err := authz.UserIDFromContext(r.Context())
		if err == nil && authedUserID != "" {
			userID = authedUserID
		} else if body.Email != "" && body.Name != "" && body.Password != "" {
			// New user registration via invitation — verify email matches
			if body.Email != inv.Email {
				respond.Error(w, http.StatusBadRequest, "email must match the invitation")
				return
			}
			newUser, err := auth.Create(r.Context(), db, auth.CreateParams{
				Email:    body.Email,
				Name:     body.Name,
				Password: body.Password,
			})
			if err != nil {
				fail(w, err)
				return
			}
			userID = newUser.ID

			// Send verification email if required
			auth.TrySendVerificationEmail(r, db, newUser.ID, newUser.Email)
		} else {
			respond.Error(w, http.StatusBadRequest, "must be authenticated or provide email, name, and password")
			return
		}

		if err := Accept(r.Context(), db, body.Token, userID); err != nil {
			fail(w, err)
			return
		}

		ws, _ := workspaces.Get(r.Context(), db, inv.WorkspaceID)

		respond.JSON(w, http.StatusOK, map[string]string{
			"status":         "accepted",
			"workspace_slug": ws.Slug,
		})
	}
}

// sendInvitationEmail renders and sends the invitation email.
func sendInvitationEmail(r *http.Request, db *sqlx.DB, rawToken string, inv Invitation, inviterUserID string) {
	ctx := r.Context()

	// Get workspace name and inviter name
	ws, _ := workspaces.Get(ctx, db, inv.WorkspaceID)
	inviter, _ := auth.Get(ctx, db, inviterUserID)

	wsName := inv.WorkspaceID
	if ws.ID != "" {
		wsName = ws.Name
	}
	inviterName := "A team member"
	if inviter.ID != "" {
		inviterName = inviter.Name
	}

	// Build accept URL
	baseURL := instance.ResolveBaseURL(ctx, db, r)
	acceptURL := fmt.Sprintf("%s/invitations/accept?token=%s", baseURL, rawToken)

	body, err := email.RenderTemplate("invitation", struct {
		WorkspaceName string
		InviterName   string
		AcceptURL     string
	}{wsName, inviterName, acceptURL})
	if err != nil {
		slog.Error("failed to render invitation email", "error", err)
		return
	}

	smtpConfig, _ := instance.LoadSMTPConfig(ctx, db)
	if err := email.Send(smtpConfig, email.Message{
		To:      inv.Email,
		Subject: fmt.Sprintf("You're invited to %s on Tookly", wsName),
		Body:    body,
	}); err != nil {
		slog.Error("failed to send invitation email", "error", err, "to", inv.Email)
	}
}
