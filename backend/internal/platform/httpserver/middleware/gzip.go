package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer       *gzip.Writer
	wroteHeader  bool
	isCompressible bool
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	if g.wroteHeader {
		return
	}
	g.wroteHeader = true

	// Check if already encoded (e.g. pre-compressed)
	if g.Header().Get("Content-Encoding") != "" {
		g.isCompressible = false
		g.ResponseWriter.WriteHeader(code)
		return
	}

	contentType := g.Header().Get("Content-Type")
	// If Content-Type is empty at header write, let it be evaluated or default to true for text/json/html
	if contentType != "" && !isCompressibleContentType(contentType) {
		g.isCompressible = false
		g.ResponseWriter.WriteHeader(code)
		return
	}

	g.isCompressible = true
	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Set("Vary", "Accept-Encoding")
	g.Header().Del("Content-Length") // Length changes after gzip compression

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(g.ResponseWriter)
	g.writer = gz

	g.ResponseWriter.WriteHeader(code)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if !g.wroteHeader {
		// Detect content type if not set yet
		if g.Header().Get("Content-Type") == "" {
			g.Header().Set("Content-Type", http.DetectContentType(b))
		}
		g.WriteHeader(http.StatusOK)
	}

	if g.isCompressible && g.writer != nil {
		return g.writer.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) Close() {
	if g.writer != nil {
		_ = g.writer.Close()
		gzipWriterPool.Put(g.writer)
		g.writer = nil
	}
}

// Gzip returns a middleware that transparently compresses HTTP responses when requested.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip compression for clients that do not support gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip compression for HEAD requests
		if r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		// Skip WebSocket upgrades
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			isCompressible: true,
		}
		defer gzw.Close()

		next.ServeHTTP(gzw, r)
	})
}

func isCompressibleContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	if idx := strings.Index(ct, ";"); idx >= 0 {
		ct = ct[:idx]
	}
	ct = strings.TrimSpace(ct)

	// Incompressible formats
	switch {
	case strings.HasPrefix(ct, "image/jpeg"),
		strings.HasPrefix(ct, "image/png"),
		strings.HasPrefix(ct, "image/webp"),
		strings.HasPrefix(ct, "image/gif"),
		strings.HasPrefix(ct, "video/"),
		strings.HasPrefix(ct, "audio/"),
		strings.HasPrefix(ct, "application/zip"),
		strings.HasPrefix(ct, "application/gzip"),
		strings.HasPrefix(ct, "application/x-rar-compressed"),
		strings.HasPrefix(ct, "application/pdf"):
		return false
	}

	// Compressible text/code/data formats
	if strings.HasPrefix(ct, "text/") ||
		strings.Contains(ct, "json") ||
		strings.Contains(ct, "javascript") ||
		strings.Contains(ct, "xml") ||
		strings.Contains(ct, "svg") ||
		strings.Contains(ct, "html") ||
		strings.Contains(ct, "css") {
		return true
	}

	return false
}
