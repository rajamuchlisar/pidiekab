package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const (
	// CSRFCookieName is the standard cookie name for CSRF tokens.
	CSRFCookieName = "pidiekab_csrf"
	// CSRFHeaderName is the request header carrying the CSRF token.
	CSRFHeaderName = "X-CSRF-Token"
)

// CSRFProtection enforces origin checking and double-submit token validation on mutating requests.
func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods do not require CSRF validation
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// Login endpoint is exempt from double-submit token because the session does not exist yet.
		// However, it is still protected by Origin verification.
		isLoginEndpoint := r.URL.Path == "/api/v1/admin/auth/login"

		// 1. Origin header defense-in-depth verification
		origin := r.Header.Get("Origin")
		if origin != "" {
			parsedOrigin, err := url.Parse(origin)
			if err != nil || parsedOrigin.Hostname() == "" {
				writeForbidden(w, r, "Invalid Origin header.")
				return
			}
			originHost := parsedOrigin.Hostname()
			reqHost := r.Host
			if strings.Contains(reqHost, ":") {
				if h, _, found := strings.Cut(reqHost, ":"); found && h != "" {
					reqHost = h
				}
			}
			// Compare origin hostname to request hostname (case-insensitive)
			if !strings.EqualFold(originHost, reqHost) {
				writeForbidden(w, r, "Cross-origin request rejected.")
				return
			}
		}

		if isLoginEndpoint {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Double-submit cookie or custom header validation for mutating requests
		reqToken := r.Header.Get(CSRFHeaderName)
		cookie, err := r.Cookie(CSRFCookieName)

		if reqToken == "" || err != nil || cookie.Value == "" {
			writeForbidden(w, r, "CSRF token missing.")
			return
		}

		if subtle.ConstantTimeCompare([]byte(reqToken), []byte(cookie.Value)) != 1 {
			writeForbidden(w, r, "CSRF token mismatch.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeForbidden(w http.ResponseWriter, r *http.Request, detail string) {
	reqID := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":       "about:blank",
		"title":      "Forbidden",
		"status":     http.StatusForbidden,
		"code":       "CSRF_ERROR",
		"detail":     detail,
		"request_id": reqID,
	})
}
