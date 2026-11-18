package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/workers/sitegeneration"
)

func handleVersionCLI() {
	fmt.Println("==================================================")
	fmt.Println("PidieKab Regional Promotion & Discovery Platform")
	fmt.Println("Version:      v1.0.0")
	fmt.Printf("Go Runtime:   %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Println("Architecture: Modular Monolith Go single binary")
	fmt.Println("Persistence:  PostgreSQL 17 sole datastore")
	fmt.Println("Frontends:    Vue 3 SSG (Public) & Vue 3 SPA (Admin CMS)")
	fmt.Println("Target:       Kabupaten Pidie, Aceh, Indonesia")
	fmt.Println("==================================================")
}

func printUsage() {
	fmt.Println("PidieKab Regional Platform Server & Operator CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  pidiekab                       Start HTTP/HTTPS application server")
	fmt.Println("  pidiekab version               Display application runtime version")
	fmt.Println("  pidiekab db status             Inspect PostgreSQL connection and table inventory")
	fmt.Println("  pidiekab ssg build             Trigger immediate manual SSG generation & release switch")
	fmt.Println("  pidiekab seed                  Seed baseline geographical locations and demo content")
	fmt.Println("  pidiekab admin list            List all registered system administrators")
	fmt.Println("  pidiekab admin create          Create a new system administrator")
	fmt.Println("  pidiekab admin reset-password  Reset an administrator password and revoke active sessions")
	fmt.Println("  pidiekab help                  Show this operator help information")
}

func handleDBStatusCLI(cfg config.Config, logger *slog.Logger) {
	if cfg.DatabaseURL == "" {
		fmt.Println("Error: PIDIEKAB_DATABASE_URL is not set.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startPing := time.Now()
	dbPool, err := database.NewPool(ctx, database.Config{
		URL: cfg.DatabaseURL,
	}, logger)
	if err != nil {
		fmt.Printf("ERROR: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		fmt.Printf("ERROR: Database ping failed: %v\n", err)
		os.Exit(1)
	}
	latency := time.Since(startPing)

	var pgVersion string
	var dbName string
	var dbSize string
	_ = dbPool.QueryRow(ctx, "SELECT version()").Scan(&pgVersion)
	_ = dbPool.QueryRow(ctx, "SELECT current_database(), pg_size_pretty(pg_database_size(current_database()))").Scan(&dbName, &dbSize)

	var tableCount int
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)

	var latestMigration int64
	var migrationApplied bool
	_ = dbPool.QueryRow(ctx, "SELECT version_id, is_applied FROM goose_db_version ORDER BY id DESC LIMIT 1").Scan(&latestMigration, &migrationApplied)

	var adminCount, catCount, contentCount, publishedCount, redirectCount, auditCount int
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM admins").Scan(&adminCount)
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&catCount)
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM contents").Scan(&contentCount)
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM content_translations WHERE status = 'published'").Scan(&publishedCount)
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM url_redirects").Scan(&redirectCount)
	_ = dbPool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&auditCount)

	fmt.Println("==================================================")
	fmt.Println("PidieKab PostgreSQL Database Status")
	fmt.Println("==================================================")
	fmt.Printf("Database Name:       %s\n", dbName)
	fmt.Printf("Database Size:       %s\n", dbSize)
	fmt.Printf("Ping Latency:        %v\n", latency)
	fmt.Printf("Public Tables:       %d\n", tableCount)
	if latestMigration > 0 {
		fmt.Printf("Latest Migration:    Goose #%d (applied: %v)\n", latestMigration, migrationApplied)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Println("Record Inventory:")
	fmt.Printf("  Administrators:    %d\n", adminCount)
	fmt.Printf("  Categories:        %d\n", catCount)
	fmt.Printf("  Total Contents:    %d\n", contentCount)
	fmt.Printf("  Published Locales: %d\n", publishedCount)
	fmt.Printf("  URL Redirects:     %d\n", redirectCount)
	fmt.Printf("  Audit Log Records: %d\n", auditCount)
	fmt.Println("--------------------------------------------------")
	if len(pgVersion) > 60 {
		pgVersion = pgVersion[:60] + "..."
	}
	fmt.Printf("Engine:              %s\n", pgVersion)
	fmt.Println("==================================================")
}

func handleSSGBuildCLI(cfg config.Config, logger *slog.Logger) {
	if cfg.DatabaseURL == "" {
		fmt.Println("Error: PIDIEKAB_DATABASE_URL is not set.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	fmt.Println("==================================================")
	fmt.Println("Starting Manual On-Demand SSG Build Pipeline")
	fmt.Println("==================================================")

	dbPool, err := database.NewPool(ctx, database.Config{
		URL: cfg.DatabaseURL,
	}, logger)
	if err != nil {
		fmt.Printf("ERROR: Database connection failed: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	releaseMgr, err := sitegeneration.NewReleaseManager(cfg.SSGOutputRoot, cfg.SSGReleasesToKeep)
	if err != nil {
		fmt.Printf("ERROR: Release manager initialization failed: %v\n", err)
		os.Exit(1)
	}

	generator := sitegeneration.NewGenerator(dbPool, "https://pidiekab.go.id")

	startTime := time.Now()
	fmt.Println("--> Creating atomic staging release directory...")
	releaseID, stagingPath, err := releaseMgr.CreateStagingDir()
	if err != nil {
		fmt.Printf("ERROR: Failed to create staging directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("--> Generating static localized HTML, sitemap, and robots.txt...")
	if err := generator.Generate(ctx, stagingPath); err != nil {
		fmt.Printf("ERROR: Site generation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("--> Activating new release symlink...")
	activePath, err := releaseMgr.ActivateRelease(releaseID, stagingPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to activate release: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	fmt.Println("==================================================")
	fmt.Printf("SSG Build SUCCESSFUL in %v\n", duration)
	fmt.Printf("Release Tag:    %s\n", releaseID)
	fmt.Printf("Active Path:    %s\n", activePath)
	fmt.Printf("Public URL:     https://pidiekab.go.id/\n")
	fmt.Println("==================================================")
}
