package sitegeneration_test

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pidiekab/internal/platform/database"
	"pidiekab/internal/workers/sitegeneration"
)

func TestReleaseManagerSafety(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "pidiekab-rm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempRoot)

	rm, err := sitegeneration.NewReleaseManager(tempRoot, 2)
	if err != nil {
		t.Fatalf("failed to create release manager: %v", err)
	}

	// 1. Initial should fail when no release has been made
	_, err = rm.GetActiveReleaseDir()
	if err == nil {
		t.Fatal("expected error getting active release before any release exists")
	}

	// 2. Create invalid staging release (missing sitemap & robots)
	releaseID, stagingDir, err := rm.CreateStagingDir()
	if err != nil {
		t.Fatalf("failed to create staging dir: %v", err)
	}

	_ = os.MkdirAll(filepath.Join(stagingDir, "id"), 0755)
	_ = os.WriteFile(filepath.Join(stagingDir, "id", "index.html"), []byte("<html>ID</html>"), 0644)

	// Validate should fail
	if err := rm.ValidateRelease(stagingDir); err == nil {
		t.Fatal("expected validation to fail due to missing en/index.html and sitemap")
	}

	// 3. Complete release files
	_ = os.MkdirAll(filepath.Join(stagingDir, "en"), 0755)
	_ = os.WriteFile(filepath.Join(stagingDir, "en", "index.html"), []byte("<html>EN</html>"), 0644)
	_ = os.WriteFile(filepath.Join(stagingDir, "sitemap.xml"), []byte("<urlset></urlset>"), 0644)
	_ = os.WriteFile(filepath.Join(stagingDir, "robots.txt"), []byte("User-agent: *"), 0644)

	activePath, err := rm.ActivateRelease(releaseID, stagingDir)
	if err != nil {
		t.Fatalf("failed to activate valid release: %v", err)
	}

	activeDir, err := rm.GetActiveReleaseDir()
	if err != nil {
		t.Fatalf("failed to get active release: %v", err)
	}
	if activeDir != activePath {
		t.Fatalf("expected active dir %s, got %s", activePath, activeDir)
	}
}

func TestSSGJobLifecycleAndGeneration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@127.0.0.1:5432/pidiekab_test?sslmode=disable"
	}
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	pool, err := database.NewPool(ctx, database.Config{
		URL:      dbURL,
		MaxConns: 5,
		MinConns: 1,
	}, logger)
	if err != nil {
		t.Skipf("skipping ssg integration test (db unavailable): %v", err)
	}
	defer pool.Close()

	tempRoot, err := os.MkdirTemp("", "pidiekab-ssg-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempRoot)

	jobRepo := sitegeneration.NewJobRepository(pool)
	releaseMgr, err := sitegeneration.NewReleaseManager(tempRoot, 3)
	if err != nil {
		t.Fatal(err)
	}

	generator := sitegeneration.NewGenerator(pool, "https://pidiekab.web.id")
	worker := sitegeneration.NewWorker(jobRepo, generator, releaseMgr, 1*time.Second, 3, logger)

	// 1. Enqueue job
	jobID, err := jobRepo.Enqueue(ctx, nil, "site.rebuild", map[string]any{"reason": "integration_test"})
	if err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}
	if jobID <= 0 {
		t.Fatalf("expected positive job id, got %d", jobID)
	}

	// 2. Process job via worker
	processed, err := worker.ProcessSingleJob(ctx)
	if err != nil {
		t.Fatalf("failed to process single job: %v", err)
	}
	if !processed {
		t.Fatal("expected worker to claim and process job")
	}

	// 3. Verify active release
	activeDir, err := releaseMgr.GetActiveReleaseDir()
	if err != nil {
		t.Fatalf("expected active release dir: %v", err)
	}

	// Verify required files
	checkFile := func(relPath string, shouldContain string) {
		full := filepath.Join(activeDir, relPath)
		data, err := os.ReadFile(full)
		if err != nil {
			t.Fatalf("missing expected generated file %s: %v", relPath, err)
		}
		if shouldContain != "" && !strings.Contains(string(data), shouldContain) {
			t.Errorf("file %s does not contain expected snippet %q", relPath, shouldContain)
		}
	}

	checkFile("id/index.html", "Bahasa Indonesia")
	checkFile("en/index.html", "English")
	checkFile("sitemap.xml", "https://pidiekab.web.id/id/")
	checkFile("robots.txt", "Sitemap: https://pidiekab.web.id/sitemap.xml")
	checkFile("index.html", "content=\"0;url=/id/\"")
}
