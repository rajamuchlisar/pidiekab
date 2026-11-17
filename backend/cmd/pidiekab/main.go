package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	categoryApp "pidiekab/internal/modules/category/application"
	categoryPostgres "pidiekab/internal/modules/category/infrastructure/postgres"
	categoryHttp "pidiekab/internal/modules/category/transport/http"
	contentApp "pidiekab/internal/modules/content/application"
	contentPostgres "pidiekab/internal/modules/content/infrastructure/postgres"
	contentHttp "pidiekab/internal/modules/content/transport/http"
	identityApp "pidiekab/internal/modules/identity/application"
	identityPostgres "pidiekab/internal/modules/identity/infrastructure/postgres"
	identityHttp "pidiekab/internal/modules/identity/transport/http"
	locationPostgres "pidiekab/internal/modules/location/infrastructure/postgres"
	locationHttp "pidiekab/internal/modules/location/transport/http"
	mediaApp "pidiekab/internal/modules/media/application"
	mediaPostgres "pidiekab/internal/modules/media/infrastructure/postgres"
	mediaStorage "pidiekab/internal/modules/media/infrastructure/storage"
	mediaHttp "pidiekab/internal/modules/media/transport/http"
	settingsApp "pidiekab/internal/modules/settings/application"
	settingsPostgres "pidiekab/internal/modules/settings/infrastructure/postgres"
	settingsHttp "pidiekab/internal/modules/settings/transport/http"
	pageApp "pidiekab/internal/modules/staticpage/application"
	pagePostgres "pidiekab/internal/modules/staticpage/infrastructure/postgres"
	pageHttp "pidiekab/internal/modules/staticpage/transport/http"
	"pidiekab/internal/platform/authn"
	"pidiekab/internal/platform/config"
	"pidiekab/internal/platform/database"
	"pidiekab/internal/platform/httpserver"
	"pidiekab/internal/workers/sitegeneration"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("fatal: invalid configuration on startup", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	// Check for CLI subcommands (e.g. `pidiekab admin create ...`)
	if len(os.Args) > 1 && os.Args[1] == "admin" {
		handleAdminCLI(cfg, logger)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var dbPool *pgxpool.Pool
	var registrars []httpserver.RouteRegistrar

	if cfg.DatabaseURL != "" {
		dbPool, err = database.NewPool(ctx, database.Config{
			URL:               cfg.DatabaseURL,
			MaxConns:          cfg.DatabaseMaxConns,
			MinConns:          cfg.DatabaseMinConns,
			MaxConnLifetime:   cfg.DatabaseMaxConnLifetime,
			MaxConnIdleTime:   cfg.DatabaseMaxConnIdleTime,
			HealthCheckPeriod: cfg.DatabaseHealthCheckPeriod,
		}, logger)
		if err != nil {
			logger.Error("fatal: database connection failed on startup", "error", err)
			os.Exit(1)
		}
		defer dbPool.Close()

		isProd := cfg.Environment == "production"

		// 1. Audit
		auditRepo := auditPostgres.NewRepository(dbPool)

		// 2. Identity & Auth
		adminRepo := identityPostgres.NewAdminRepository(dbPool)
		sessionRepo := identityPostgres.NewSessionRepository(dbPool)
		authService := identityApp.NewAuthService(
			adminRepo,
			sessionRepo,
			auditRepo,
			authn.DefaultArgon2idParams(),
			12*time.Hour,
			logger,
		)
		identityHandler := identityHttp.NewHandler(authService, isProd)
		registrars = append(registrars, identityHandler)

		// 3. Location Master
		locationRepo := locationPostgres.NewRepository(dbPool)
		if err := locationRepo.SeedPidieBaselineIfEmpty(ctx); err != nil {
			logger.Warn("location baseline seed skipped or failed", "error", err)
		}
		locationHandler := locationHttp.NewHandler(locationRepo, authService)
		registrars = append(registrars, locationHandler)

		// 4. Categories
		categoryRepo := categoryPostgres.NewRepository(dbPool)
		categoryService := categoryApp.NewService(categoryRepo, auditRepo)
		categoryHandler := categoryHttp.NewHandler(categoryService, authService)
		registrars = append(registrars, categoryHandler)

		// 5. Contents, Translations & Sources
		contentRepo := contentPostgres.NewContentRepository(dbPool)
		translationRepo := contentPostgres.NewTranslationRepository(dbPool)
		sourceRepo := contentPostgres.NewSourceRepository(dbPool)
		contentService := contentApp.NewContentService(contentRepo, translationRepo, sourceRepo, categoryService, auditRepo)
		contentHandler := contentHttp.NewHandler(contentService, authService)
		registrars = append(registrars, contentHandler)

		// 6. Media Library & Secure Storage
		mediaStoragePath := os.Getenv("PIDIEKAB_STORAGE_PATH")
		if mediaStoragePath == "" {
			mediaStoragePath = "./data/media"
		}
		localStorage, err := mediaStorage.NewLocalStorage(mediaStoragePath)
		if err != nil {
			logger.Error("failed to initialize media storage", "path", mediaStoragePath, "error", err)
			os.Exit(1)
		}
		mediaRepo := mediaPostgres.NewMediaRepository(dbPool)
		mediaService := mediaApp.NewMediaService(mediaRepo, localStorage, auditRepo)
		mediaHandler := mediaHttp.NewMediaHandler(mediaService, authService)
		registrars = append(registrars, mediaHandler)

		// 7. Static Information Pages
		pageRepo := pagePostgres.NewPageRepository(dbPool)
		pageService := pageApp.NewPageService(pageRepo, auditRepo)
		pageHandler := pageHttp.NewPageHandler(pageService, authService)
		registrars = append(registrars, pageHandler)

		// 8. Website Settings
		settingsRepo := settingsPostgres.NewSettingsRepository(dbPool)
		settingsService := settingsApp.NewSettingsService(settingsRepo, auditRepo)
		settingsHandler := settingsHttp.NewSettingsHandler(settingsService, authService)
		registrars = append(registrars, settingsHandler)

		// 9. Site Generation Workers & SSG Pipeline
		releaseMgr, err := sitegeneration.NewReleaseManager(cfg.SSGOutputRoot, cfg.SSGReleasesToKeep)
		if err != nil {
			logger.Error("failed to initialize SSG release manager", "error", err)
			os.Exit(1)
		}
		registrars = append(registrars, releaseMgr)

		jobRepo := sitegeneration.NewJobRepository(dbPool)
		contentService.SetJobEnqueuer(jobRepo)
		pageService.SetJobEnqueuer(jobRepo)

		generator := sitegeneration.NewGenerator(dbPool, "https://pidiekab.go.id")
		ssgWorker := sitegeneration.NewWorker(jobRepo, generator, releaseMgr, cfg.SSGWorkerPollInterval, cfg.SSGJobMaxAttempts, logger)

		// Run background SSG worker lifecycle
		go ssgWorker.Start(ctx)

		// Trigger initial build if no release exists yet
		if _, err := releaseMgr.GetActiveReleaseDir(); err != nil {
			logger.Info("no active SSG release found on startup; scheduling initial site generation")
			_, _ = jobRepo.Enqueue(ctx, nil, "site.initial_build", map[string]any{"trigger": "startup"})
		}
	} else {
		logger.Warn("PIDIEKAB_DATABASE_URL is not configured; running without database connection pool")
	}

	server := httpserver.New(cfg, logger, dbPool, registrars...)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting pidiekab server",
			"addr", cfg.Address,
			"environment", cfg.Environment,
			"log_level", cfg.LogLevel.String(),
		)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received, commencing graceful shutdown", "timeout", cfg.ShutdownTimeout.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("server shutdown completed successfully")
	}
}

func handleAdminCLI(cfg config.Config, logger *slog.Logger) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: pidiekab admin <subcommand> [flags]")
		fmt.Println("Subcommands: create")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "create", "bootstrap":
		cmd := flag.NewFlagSet("admin create", flag.ExitOnError)
		emailFlag := cmd.String("email", "", "Admin email address")
		nameFlag := cmd.String("name", "", "Admin display name")
		passwordFlag := cmd.String("password", "", "Admin password (minimum 8 characters)")

		if err := cmd.Parse(os.Args[3:]); err != nil {
			fmt.Println("Failed to parse flags:", err)
			os.Exit(1)
		}

		if *emailFlag == "" || *nameFlag == "" || *passwordFlag == "" {
			fmt.Println("Error: --email, --name, and --password are required.")
			cmd.Usage()
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dbPool, err := database.NewPool(ctx, database.Config{
			URL: cfg.DatabaseURL,
		}, logger)
		if err != nil {
			logger.Error("database connection failed", "error", err)
			os.Exit(1)
		}
		defer dbPool.Close()

		adminRepo := identityPostgres.NewAdminRepository(dbPool)
		sessionRepo := identityPostgres.NewSessionRepository(dbPool)
		auditRepo := auditPostgres.NewRepository(dbPool)

		authService := identityApp.NewAuthService(
			adminRepo,
			sessionRepo,
			auditRepo,
			authn.DefaultArgon2idParams(),
			12*time.Hour,
			logger,
		)

		admin, err := authService.BootstrapAdmin(ctx, *emailFlag, *nameFlag, *passwordFlag)
		if err != nil {
			logger.Error("failed to create admin", "error", err)
			os.Exit(1)
		}

		logger.Info("administrator created successfully", "id", admin.ID, "email", admin.Email, "name", admin.DisplayName)
		fmt.Printf("Administrator %q (ID %d) created successfully.\n", admin.Email, admin.ID)
	default:
		fmt.Printf("Unknown admin subcommand %q. Available: create\n", os.Args[2])
		os.Exit(1)
	}
}
