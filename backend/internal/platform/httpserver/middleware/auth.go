package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"pidiekab/internal/modules/identity/domain"
)

const (
	// AdminContextKey is the context key for the authenticated administrator.
	AdminContextKey contextKey = "pidiekab_admin"
	// SessionContextKey is the context key for the authenticated session.
	SessionContextKey contextKey = "pidiekab_session"
	// AdminSessionCookieName is the standard cookie name for admin authentication.
	AdminSessionCookieName = "pidiekab_admin_session"
)

// SessionAuthenticator defines the contract for validating an admin session token.
type SessionAuthenticator interface {
	AuthenticateSession(ctx context.Context, rawToken string) (*domain.Admin, *domain.Session, error)
}

// RequireAdminSession is a middleware that enforces active administrator authentication.
func RequireAdminSession(auth SessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AdminSessionCookieName)
			if err != nil || cookie.Value == "" {
				writeUnauthorized(w, r, "Authentication required.")
				return
			}

			admin, sess, err := auth.AuthenticateSession(r.Context(), cookie.Value)
			if err != nil {
				writeUnauthorized(w, r, "Invalid or expired session.")
				return
			}

			ctx := context.WithValue(r.Context(), AdminContextKey, admin)
			ctx = context.WithValue(ctx, SessionContextKey, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminFromContext retrieves the authenticated administrator from request context.
func AdminFromContext(ctx context.Context) (*domain.Admin, bool) {
	admin, ok := ctx.Value(AdminContextKey).(*domain.Admin)
	return admin, ok && admin != nil
}

// SessionFromContext retrieves the authenticated session from request context.
func SessionFromContext(ctx context.Context) (*domain.Session, bool) {
	sess, ok := ctx.Value(SessionContextKey).(*domain.Session)
	return sess, ok && sess != nil
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	reqID := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":       "about:blank",
		"title":      "Unauthorized",
		"status":     http.StatusUnauthorized,
		"code":       "UNAUTHORIZED",
		"detail":     detail,
		"request_id": reqID,
	})
}
