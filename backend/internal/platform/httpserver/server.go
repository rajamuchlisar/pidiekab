package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/httpserver/middleware"
)

// DatabasePinger is an interface for checking database health.
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

// BuildHandler builds the router, handlers, and attaches middleware.
func BuildHandler(cfg config.Config, logger *slog.Logger, db DatabasePinger) http.Handler {
	mux := http.NewServeMux()

	// 1. Health endpoints
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

	// 2. Root redirect: / -> /id/
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/id/", http.StatusMovedPermanently)
	})

	// 3. Static serving for Admin CMS (/cms/ and /cms/*)
	mux.HandleFunc("/cms/", func(w http.ResponseWriter, r *http.Request) {
		serveCMSStatic(w, r, cfg.CMSStaticDir)
	})

	// 4. Public static serving (SSG generated HTML, assets, sitemap.xml, robots.txt)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		servePublicStatic(w, r, cfg.PublicStaticDir)
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
func New(cfg config.Config, logger *slog.Logger, db DatabasePinger) *http.Server {
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           BuildHandler(cfg, logger, db),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
}

func servePublicStatic(w http.ResponseWriter, r *http.Request, rootDir string) {
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
