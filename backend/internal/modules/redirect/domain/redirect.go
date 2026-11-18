package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyPath    = errors.New("old_path and new_path cannot be empty")
	ErrSelfRedirect = errors.New("old_path cannot be identical to new_path")
)

// Redirect represents a permanent 301 URL redirect mapping from old_path to new_path.
type Redirect struct {
	ID         int64
	OldPath    string
	NewPath    string
	StatusCode int
	CreatedBy  *int64
	CreatedAt  time.Time
}

// NormalizePath trims whitespace, ensures a leading slash, and canonicalizes directory paths.
func NormalizePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// Canonicalize directory style URL if not referencing a static file with an extension
	lastSegment := p
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		lastSegment = p[idx+1:]
	}
	if !strings.Contains(lastSegment, ".") && !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	return p
}
