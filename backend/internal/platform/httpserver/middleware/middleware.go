package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

type contextKey string

const (
	RequestIDHeader                = "X-Request-Id"
	requestIDContextKey contextKey = "pidiekab.request_id"
)

var validRequestIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,64}$`)

// ContextWithRequestID attaches a request ID to the given context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// RequestIDFromContext retrieves the request ID from context if present.
func RequestIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(requestIDContextKey).(string); ok {
		return val
	}
	return ""
}

// RequestID middleware ensures every request has a validated or newly generated X-Request-Id.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDHeader)
		if !validRequestIDRegex.MatchString(reqID) {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				reqID = fmt.Sprintf("fallback-%d", time.Now().UnixNano())
			} else {
				reqID = hex.EncodeToString(b)
			}
		}

		ctx := ContextWithRequestID(r.Context(), reqID)
		w.Header().Set(RequestIDHeader, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SecurityHeaders applies production baseline HTTP security headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		// Media assets & PDFs require relaxed frame & opener policies so browser PDF plugins and media players can render
		isMedia := strings.HasPrefix(r.URL.Path, "/media/") || strings.HasSuffix(strings.ToLower(r.URL.Path), ".pdf")
		if isMedia {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Content-Security-Policy", "default-src 'self' blob:; object-src 'self' blob:; img-src 'self' data: https: blob:; frame-ancestors 'self'")
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		} else {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; object-src 'none'; script-src 'self' 'unsafe-inline' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: blob:; font-src 'self'; connect-src 'self' https://cloudflareinsights.com; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		}

		next.ServeHTTP(w, r)
	})
}

// responseRecorder captures the status code and bytes written for logging.
type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

func (rw *responseRecorder) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Logger creates access logging middleware using slog.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			reqID := RequestIDFromContext(r.Context())

			logger.Info("http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes", rec.bytesWritten,
				"request_id", reqID,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// Recovery catches panics, logs stack trace with request ID, and writes a generic 500 JSON.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					reqID := RequestIDFromContext(r.Context())
					stack := string(debug.Stack())

					logger.Error("panic recovered during http request",
						"error", fmt.Sprintf("%v", rec),
						"request_id", reqID,
						"path", r.URL.Path,
						"stack", stack,
					)

					w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"type":       "about:blank",
						"title":      "Internal Server Error",
						"status":     http.StatusInternalServerError,
						"detail":     "An unexpected error occurred while processing your request.",
						"request_id": reqID,
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
