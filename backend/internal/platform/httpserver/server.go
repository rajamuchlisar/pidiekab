package httpserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/httpserver/middleware"
)

// DatabasePinger is an interface for checking database health.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// ReleaseLocator provides dynamic resolution of the active SSG release directory.
type ReleaseLocator interface {
	GetActiveReleaseDir() (string, error)
}

// RedirectFinder provides lookup for registered URL redirects.
type RedirectFinder interface {
	FindRedirect(ctx context.Context, path string) (targetPath string, statusCode int, err error)
}

// RouteRegistrar allows domain modules to attach their HTTP routes onto the ServeMux.
type RouteRegistrar interface {
	RegisterRoutes(mux *http.ServeMux)
}

// BuildHandler builds the router, handlers, and attaches middleware.
func BuildHandler(cfg config.Config, logger *slog.Logger, db DatabasePinger, registrars ...RouteRegistrar) http.Handler {
	mux := http.NewServeMux()
	startTime := time.Now()

	// 1. Health & Observability endpoints
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "ok",
			"service":     "pidiekab",
			"environment": cfg.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":      "not_ready",
				"reason":      "database_not_configured",
				"service":     "pidiekab",
				"environment": cfg.Environment,
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(pingCtx); err != nil {
			logger.Warn("readiness probe failed: database ping error", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":      "not_ready",
				"reason":      "database_unreachable",
				"service":     "pidiekab",
				"environment": cfg.Environment,
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "ok",
			"service":     "pidiekab",
			"database":    "connected",
			"environment": cfg.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 2. Metrics endpoint
	mux.HandleFunc("GET /health/metrics", func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		dbStatus := "unconfigured"
		if db != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
			defer cancel()
			if err := db.Ping(pingCtx); err == nil {
				dbStatus = "connected"
			} else {
				dbStatus = "unreachable"
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":         "ok",
			"service":        "pidiekab",
			"uptime_seconds": int64(time.Since(startTime).Seconds()),
			"database":       dbStatus,
			"goroutines":     runtime.NumGoroutine(),
			"mem_alloc_mb":   float64(mem.Alloc) / (1024 * 1024),
			"mem_sys_mb":     float64(mem.Sys) / (1024 * 1024),
			"timestamp":      time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 3. Module routes (Identity, Content, Redirects, etc.)
	var releaseLocator ReleaseLocator
	var redirectFinder RedirectFinder
	for _, reg := range registrars {
		if reg != nil {
			reg.RegisterRoutes(mux)
			if rl, ok := reg.(ReleaseLocator); ok {
				releaseLocator = rl
			}
			if rf, ok := reg.(RedirectFinder); ok {
				redirectFinder = rf
			}
		}
	}

	// 4. Root redirect: / -> /id/
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/id/", http.StatusMovedPermanently)
	})

	// 5. Static serving for Admin CMS (/cms/ and /cms/*)
	mux.HandleFunc("/cms/", func(w http.ResponseWriter, r *http.Request) {
		serveCMSStatic(w, r, cfg.CMSStaticDir)
	})

	// 6. Public static serving (SSG generated HTML, assets, sitemap.xml, robots.txt)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rootDir := cfg.PublicStaticDir
		if releaseLocator != nil {
			if activeDir, err := releaseLocator.GetActiveReleaseDir(); err == nil && activeDir != "" {
				rootDir = activeDir
			}
		}
		servePublicStatic(w, r, rootDir, redirectFinder)
	})

	// Standard middleware wrapping order:
	// RequestID -> Recovery -> Logger -> SecurityHeaders -> ServeMux
	var handler http.Handler = mux
	handler = middleware.SecurityHeaders(handler)
	handler = middleware.Logger(logger)(handler)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.RequestID(handler)

	return handler
}

// New creates and configures an http.Server with explicit timeouts, database pinger, and middleware.
// If TLS certificates are configured in cfg, modern TLS 1.2+ configuration is applied.
func New(cfg config.Config, logger *slog.Logger, db DatabasePinger, registrars ...RouteRegistrar) *http.Server {
	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           BuildHandler(cfg, logger, db, registrars...),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{
				tls.X25519,
				tls.CurveP256,
			},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
	}

	return srv
}

// NewHTTPToHTTPSRedirectServer creates an HTTP server that unconditionally redirects all traffic to HTTPS.
func NewHTTPToHTTPSRedirectServer(addr string, logger *slog.Logger) *http.Server {
	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
			host = host[:colonIdx]
		}
		target := "https://" + host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})

	return &http.Server{
		Addr:              addr,
		Handler:           redirectHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

func servePublicStatic(w http.ResponseWriter, r *http.Request, rootDir string, redirectFinder RedirectFinder) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cleanPath := filepath.Clean(filepath.FromSlash(r.URL.Path))
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetPath := filepath.Join(rootDir, cleanPath)

	// Check if file directly exists
	stat, err := os.Stat(targetPath)
	if err == nil && !stat.IsDir() {
		applyCacheHeaders(w, r.URL.Path)
		http.ServeFile(w, r, targetPath)
		return
	}

	// Check if directory index.html exists
	indexPath := filepath.Join(targetPath, "index.html")
	if stat, err := os.Stat(indexPath); err == nil && !stat.IsDir() {
		if !strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		http.ServeFile(w, r, indexPath)
		return
	}

	// Check registered URL redirects before falling back to 404
	if redirectFinder != nil {
		if target, statusCode, err := redirectFinder.FindRedirect(r.Context(), r.URL.Path); err == nil && target != "" {
			if statusCode == 0 {
				statusCode = http.StatusMovedPermanently
			}
			http.Redirect(w, r, target, statusCode)
			return
		}
	}

	http.NotFound(w, r)
}

func serveCMSStatic(w http.ResponseWriter, r *http.Request, rootDir string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	relPath := strings.TrimPrefix(r.URL.Path, "/cms/")
	cleanPath := filepath.Clean(filepath.FromSlash(relPath))
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	targetPath := filepath.Join(rootDir, cleanPath)
	stat, err := os.Stat(targetPath)
	if err == nil && !stat.IsDir() {
		applyCacheHeaders(w, r.URL.Path)
		http.ServeFile(w, r, targetPath)
		return
	}

	// SPA fallback: serve /cms/index.html with no-cache
	spaIndex := filepath.Join(rootDir, "index.html")
	if _, err := os.Stat(spaIndex); err == nil {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		http.ServeFile(w, r, spaIndex)
		return
	}

	http.NotFound(w, r)
}

func applyCacheHeaders(w http.ResponseWriter, requestPath string) {
	if strings.HasPrefix(requestPath, "/assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if strings.HasSuffix(requestPath, ".html") || strings.HasSuffix(requestPath, "/") {
		w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
