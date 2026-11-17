package sitegeneration

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrIncompleteRelease = errors.New("release validation failed: missing essential files")
	ErrNoActiveRelease   = errors.New("no active SSG release found")
)

type ReleaseManager struct {
	outputRoot     string
	releasesToKeep int
	mu             sync.RWMutex
	activeDir      string
}

func NewReleaseManager(outputRoot string, releasesToKeep int) (*ReleaseManager, error) {
	absRoot, err := filepath.Abs(outputRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SSG output root: %w", err)
	}

	if releasesToKeep < 1 {
		releasesToKeep = 5
	}

	releasesDir := filepath.Join(absRoot, "releases")
	stagingDir := filepath.Join(absRoot, "staging")
	if err := os.MkdirAll(releasesDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return nil, err
	}

	rm := &ReleaseManager{
		outputRoot:     absRoot,
		releasesToKeep: releasesToKeep,
	}

	// Try reading existing active release pointer if present
	pointerFile := filepath.Join(absRoot, "active_release.txt")
	if data, err := os.ReadFile(pointerFile); err == nil {
		active := strings.TrimSpace(string(data))
		if info, err := os.Stat(active); err == nil && info.IsDir() {
			rm.activeDir = active
		}
	}

	return rm, nil
}

func (rm *ReleaseManager) CreateStagingDir() (string, string, error) {
	releaseID := time.Now().UTC().Format("20060102_150405_000000")
	stagingPath := filepath.Join(rm.outputRoot, "staging", releaseID)
	if err := os.MkdirAll(stagingPath, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create staging dir: %w", err)
	}
	return releaseID, stagingPath, nil
}

func (rm *ReleaseManager) ValidateRelease(stagingPath string) error {
	requiredFiles := []string{
		filepath.Join("id", "index.html"),
		filepath.Join("en", "index.html"),
		"sitemap.xml",
		"robots.txt",
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join(stagingPath, file)
		info, err := os.Stat(fullPath)
		if err != nil {
			return fmt.Errorf("%w: %s is missing", ErrIncompleteRelease, file)
		}
		if info.Size() == 0 {
			return fmt.Errorf("%w: %s is empty", ErrIncompleteRelease, file)
		}
	}

	return nil
}

func (rm *ReleaseManager) ActivateRelease(releaseID, stagingPath string) (string, error) {
	if err := rm.ValidateRelease(stagingPath); err != nil {
		return "", err
	}

	releasesDir := filepath.Join(rm.outputRoot, "releases")
	targetReleasePath := filepath.Join(releasesDir, releaseID)

	// Move staging to releases
	if err := os.Rename(stagingPath, targetReleasePath); err != nil {
		return "", fmt.Errorf("failed to promote staging release to releases dir: %w", err)
	}

	// Update pointer file atomically
	pointerFile := filepath.Join(rm.outputRoot, "active_release.txt")
	tempPointer := filepath.Join(rm.outputRoot, "active_release.tmp")
	if err := os.WriteFile(tempPointer, []byte(targetReleasePath), 0644); err != nil {
		return "", fmt.Errorf("failed to write release pointer: %w", err)
	}
	if err := os.Rename(tempPointer, pointerFile); err != nil {
		return "", fmt.Errorf("failed to atomically update release pointer: %w", err)
	}

	rm.mu.Lock()
	rm.activeDir = targetReleasePath
	rm.mu.Unlock()

	// Cleanup old releases in background
	go rm.cleanupOldReleases()

	return targetReleasePath, nil
}

func (rm *ReleaseManager) GetActiveReleaseDir() (string, error) {
	rm.mu.RLock()
	dir := rm.activeDir
	rm.mu.RUnlock()

	if dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
	}

	// Try reading pointer file
	pointerFile := filepath.Join(rm.outputRoot, "active_release.txt")
	if data, err := os.ReadFile(pointerFile); err == nil {
		active := strings.TrimSpace(string(data))
		if info, err := os.Stat(active); err == nil && info.IsDir() {
			rm.mu.Lock()
			rm.activeDir = active
			rm.mu.Unlock()
			return active, nil
		}
	}

	return "", ErrNoActiveRelease
}

// RegisterRoutes fulfills the RouteRegistrar interface for automatic release locator discovery.
func (rm *ReleaseManager) RegisterRoutes(mux *http.ServeMux) {
}

func (rm *ReleaseManager) cleanupOldReleases() {
	releasesDir := filepath.Join(rm.outputRoot, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return
	}

	var releaseDirs []string
	for _, e := range entries {
		if e.IsDir() {
			releaseDirs = append(releaseDirs, filepath.Join(releasesDir, e.Name()))
		}
	}

	if len(releaseDirs) <= rm.releasesToKeep {
		return
	}

	sort.Strings(releaseDirs) // oldest first
	toRemoveCount := len(releaseDirs) - rm.releasesToKeep

	rm.mu.RLock()
	currentActive := rm.activeDir
	rm.mu.RUnlock()

	for i := 0; i < toRemoveCount; i++ {
		dir := releaseDirs[i]
		if dir == currentActive {
			continue // Never delete active release
		}
		_ = os.RemoveAll(dir)
	}
}
