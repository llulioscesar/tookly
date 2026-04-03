// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1
// Use of this software is governed by the Business Source License 1.1
// included in the LICENSE file at the root of this repository.

package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/start-codex/tookly/internal/authz"
	"github.com/start-codex/tookly/internal/email"
	"github.com/start-codex/tookly/internal/respond"
	"github.com/start-codex/tookly/internal/sessions"
)

func RegisterRoutes(mux *http.ServeMux, db *sqlx.DB) {
	// User routes
	mux.HandleFunc("POST /users", handleCreate(db))
	mux.HandleFunc("GET /users/{userID}", handleGet(db))
	// Auth routes
	mux.HandleFunc("POST /auth/login", handleLogin(db))
	mux.HandleFunc("GET /auth/me", handleMe(db))
	mux.HandleFunc("POST /auth/logout", handleLogout(db))
	mux.HandleFunc("POST /auth/change-password", handleChangePassword(db))
	// Email verification routes
	mux.HandleFunc("POST /auth/verify-email", handleVerifyEmail(db))
	mux.HandleFunc("POST /auth/resend-verification", handleResendVerification(db))
	// Password reset routes
	mux.HandleFunc("POST /auth/forgot-password", handleForgotPassword(db))
	mux.HandleFunc("POST /auth/reset-password", handleResetPassword(db))
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		respond.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrDuplicateEmail):
		respond.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidCredentials):
		respond.Error(w, http.StatusUnauthorized, err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func setSessionCookie(w http.ResponseWriter, rawToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    rawToken,
		Path:     "/",
		MaxAge:   604800,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   os.Getenv("SECURE_COOKIES") == "true",
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func handleCreate(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		params := CreateParams{Email: body.Email, Name: body.Name, Password: body.Password}
		if err := params.Validate(); err != nil {
			respond.Error(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		user, err := Create(r.Context(), db, params)
		if err != nil {
			fail(w, err)
			return
		}

		TrySendVerificationEmail(r, db, user.ID, user.Email)

		respond.JSON(w, http.StatusCreated, user)
	}
}

func handleLogin(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		user, err := Authenticate(r.Context(), db, body.Email, body.Password)
		if err != nil {
			fail(w, err)
			return
		}
		if user.ArchivedAt != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		result, err := sessions.Create(r.Context(), db, user.ID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		setSessionCookie(w, result.RawToken)
		respond.JSON(w, http.StatusOK, user)
	}
}

func handleMe(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
			return
		}
		session, err := sessions.Validate(r.Context(), db, cookie.Value)
		if err != nil {
			if sessions.IsAuthError(err) {
				clearSessionCookie(w)
				respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		user, err := Get(r.Context(), db, session.UserID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				clearSessionCookie(w)
				respond.JSON(w, http.StatusOK, map[string]any{"authenticated": false})
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		verificationRequired, _ := IsVerificationRequired(r.Context(), db)
		respond.JSON(w, http.StatusOK, map[string]any{
			"authenticated":               true,
			"user":                        user,
			"email_verification_required": verificationRequired,
		})
	}
}

func handleLogout(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie.Value != "" {
			_ = sessions.Delete(r.Context(), db, cookie.Value)
		}
		clearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleChangePassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		var body struct {
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if err := ChangePassword(r.Context(), db, userID, body.CurrentPassword, body.NewPassword); err != nil {
			if errors.Is(err, ErrInvalidCredentials) {
				respond.Error(w, http.StatusUnauthorized, "current password is incorrect")
				return
			}
			if errors.Is(err, ErrPasswordTooShort) {
				respond.Error(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			fail(w, err)
			return
		}
		// Invalidate all other sessions, preserving the current one
		cookie, _ := r.Cookie("session_id")
		if cookie != nil && cookie.Value != "" {
			_ = sessions.DeleteByUserID(r.Context(), db, userID, cookie.Value)
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "password_changed"})
	}
}

func handleGet(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authedUserID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if authedUserID != r.PathValue("userID") {
			respond.Error(w, http.StatusForbidden, "access denied")
			return
		}
		user, err := Get(r.Context(), db, r.PathValue("userID"))
		if err != nil {
			fail(w, err)
			return
		}
		respond.JSON(w, http.StatusOK, user)
	}
}

func handleVerifyEmail(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Token == "" {
			respond.Error(w, http.StatusBadRequest, "token is required")
			return
		}
		if err := VerifyEmail(r.Context(), db, body.Token); err != nil {
			if errors.Is(err, ErrVerifyTokenNotFound) || errors.Is(err, ErrVerifyTokenExpired) || errors.Is(err, ErrVerifyTokenUsed) {
				respond.Error(w, http.StatusBadRequest, "invalid_or_expired_token")
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "verified"})
	}
}

func handleResendVerification(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authz.UserIDFromContext(r.Context())
		if err != nil {
			respond.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		user, err := Get(r.Context(), db, userID)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		// Idempotent: already verified → success without sending
		if user.EmailVerifiedAt != nil {
			respond.JSON(w, http.StatusOK, map[string]string{"status": "already_verified"})
			return
		}
		baseURL := resolveBaseURL(r.Context(), db, r)
		if err := SendVerificationEmail(r.Context(), db, userID, user.Email, baseURL); err != nil {
			respond.Error(w, http.StatusInternalServerError, "failed to send verification email")
			return
		}
		respond.JSON(w, http.StatusOK, map[string]string{"status": "sent"})
	}
}

func handleForgotPassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email string `json:"email"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		// Always return 200 — no email enumeration
		user, err := GetByEmail(r.Context(), db, body.Email)
		if err != nil || user.ArchivedAt != nil {
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		rawToken, err := CreateResetToken(r.Context(), db, user.ID)
		if err != nil {
			slog.Error("failed to create reset token", "error", err, "email", body.Email)
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		baseURL := resolveBaseURL(r.Context(), db, r)
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, rawToken)

		emailBody, err := email.RenderTemplate("password_reset", struct{ ResetURL string }{resetURL})
		if err != nil {
			slog.Error("failed to render reset email template", "error", err)
			respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		smtpConfig, _ := loadSMTPConfig(r.Context(), db)
		if err := email.Send(smtpConfig, email.Message{
			To:      user.Email,
			Subject: "Reset your Tookly password",
			Body:    emailBody,
		}); err != nil {
			slog.Error("failed to send reset email", "error", err, "to", user.Email)
		}

		respond.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleResetPassword(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token       string `json:"token"`
			NewPassword string `json:"new_password"`
		}
		if err := respond.Decode(r, &body); err != nil {
			respond.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if body.Token == "" || body.NewPassword == "" {
			respond.Error(w, http.StatusBadRequest, "token and new_password are required")
			return
		}

		if err := ResetPassword(r.Context(), db, body.Token, body.NewPassword); err != nil {
			if errors.Is(err, ErrResetTokenNotFound) || errors.Is(err, ErrResetTokenExpired) || errors.Is(err, ErrResetTokenUsed) {
				respond.Error(w, http.StatusBadRequest, "invalid_or_expired_token")
				return
			}
			if errors.Is(err, ErrPasswordTooShort) {
				respond.Error(w, http.StatusUnprocessableEntity, err.Error())
				return
			}
			respond.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}

		respond.JSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
	}
}

// TrySendVerificationEmail checks if email verification is required and sends
// the verification email if so. Errors are logged but never fail the caller.
func TrySendVerificationEmail(r *http.Request, db *sqlx.DB, userID, userEmail string) {
	ctx := r.Context()
	required, err := IsVerificationRequired(ctx, db)
	if err != nil || !required {
		return
	}
	baseURL := resolveBaseURL(ctx, db, r)
	if err := SendVerificationEmail(ctx, db, userID, userEmail, baseURL); err != nil {
		slog.Error("failed to send verification email", "error", err, "user_id", userID)
	}
}
