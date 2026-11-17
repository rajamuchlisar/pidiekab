package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientRecord struct {
	count     int
	resetTime time.Time
}

// MemoryRateLimiter implements a simple sliding window limiter per client IP.
type MemoryRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientRecord
	limit   int
	window  time.Duration
}

// NewMemoryRateLimiter creates a new rate limiter with a specified request limit per time window.
func NewMemoryRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
	limiter := &MemoryRateLimiter{
		clients: make(map[string]*clientRecord),
		limit:   limit,
		window:  window,
	}

	// Periodically purge stale records
	go func() {
		ticker := time.NewTicker(2 * window)
		for now := range ticker.C {
			limiter.mu.Lock()
			for ip, rec := range limiter.clients {
				if now.After(rec.resetTime) {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

// LimitHandler wraps an http.Handler with IP-based rate limiting.
func (l *MemoryRateLimiter) LimitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		l.mu.Lock()
		now := time.Now()
		rec, exists := l.clients[ip]
		if !exists || now.After(rec.resetTime) {
			rec = &clientRecord{
				count:     1,
				resetTime: now.Add(l.window),
			}
			l.clients[ip] = rec
		} else {
			rec.count++
		}

		if rec.count > l.limit {
			l.mu.Unlock()
			reqID := r.Header.Get("X-Request-Id")
			w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":       "about:blank",
				"title":      "Too Many Requests",
				"status":     http.StatusTooManyRequests,
				"code":       "RATE_LIMITED",
				"detail":     "Too many requests. Please try again later.",
				"request_id": reqID,
			})
			return
		}
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
