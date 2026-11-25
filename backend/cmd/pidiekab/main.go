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

	auditPostgres "pidiekab/internal/modules/audit/infrastructure/postgres"
	auditHttp "pidiekab/internal/modules/audit/transport/http"
	analyticsApp "pidiekab/internal/modules/analytics/application"
	analyticsPostgres "pidiekab/internal/modules/analytics/infrastructure/postgres"
	analyticsHttp "pidiekab/internal/modules/analytics/transport/http"
	categoryApp "pidiekab/internal/modules/category/application"
	categoryPostgres "pidiekab/internal/modules/category/infrastructure/postgres"
	categoryHttp "pidiekab/internal/modules/category/transport/http"
	contentApp "pidiekab/internal/modules/content/application"
	contentPostgres "pidiekab/internal/modules/content/infrastructure/postgres"
	contentHttp "pidiekab/internal/modules/content/transport/http"
	curationApp "pidiekab/internal/modules/curation/application"
	curationPostgres "pidiekab/internal/modules/curation/infrastructure/postgres"
	curationHttp "pidiekab/internal/modules/curation/transport/http"
	downloadApp "pidiekab/internal/modules/download/application"
	downloadPostgres "pidiekab/internal/modules/download/infrastructure/postgres"
	downloadHttp "pidiekab/internal/modules/download/transport/http"
	identityApp "pidiekab/internal/modules/identity/application"
	identityPostgres "pidiekab/internal/modules/identity/infrastructure/postgres"
	identityHttp "pidiekab/internal/modules/identity/transport/http"
	layoutApp "pidiekab/internal/modules/layout/application"
	layoutPostgres "pidiekab/internal/modules/layout/infrastructure/postgres"
	layoutHttp "pidiekab/internal/modules/layout/transport/http"
	locationPostgres "pidiekab/internal/modules/location/infrastructure/postgres"
	locationHttp "pidiekab/internal/modules/location/transport/http"
	mediaApp "pidiekab/internal/modules/media/application"
	mediaPostgres "pidiekab/internal/modules/media/infrastructure/postgres"
	mediaStorage "pidiekab/internal/modules/media/infrastructure/storage"
	mediaHttp "pidiekab/internal/modules/media/transport/http"
	menuApp "pidiekab/internal/modules/menu/application"
	menuPostgres "pidiekab/internal/modules/menu/infrastructure/postgres"
	menuHttp "pidiekab/internal/modules/menu/transport/http"
	redirectPostgres "pidiekab/internal/modules/redirect/infrastructure/postgres"
	redirectHttp "pidiekab/internal/modules/redirect/transport/http"
	searchApp "pidiekab/internal/modules/search/application"
	searchPostgres "pidiekab/internal/modules/search/infrastructure/postgres"
	searchHttp "pidiekab/internal/modules/search/transport/http"
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

	"github.com/jackc/pgx/v5/pgxpool"
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

	// Check for CLI subcommands (e.g. `pidiekab admin`, `pidiekab seed`, `pidiekab ssg build`, etc.)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "-v", "--version":
			handleVersionCLI()
			return
		case "admin":
			handleAdminCLI(cfg, logger)
			return
		case "seed":
			handleSeedCLI(cfg, logger)
			return
		case "seed-drafts":
			handleSeedDraftsCLI(cfg, logger)
			return
		case "db":
			if len(os.Args) > 2 && os.Args[2] == "status" {
				handleDBStatusCLI(cfg, logger)
				return
			}
			fmt.Println("Usage: pidiekab db status")
			os.Exit(1)
		case "ssg":
			if len(os.Args) > 2 && os.Args[2] == "build" {
				handleSSGBuildCLI(cfg, logger)
				return
			}
			fmt.Println("Usage: pidiekab ssg build")
			os.Exit(1)
		case "help", "-h", "--help":
			printUsage()
			return
		}
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

		// 8b. Navigation Menus
		menuRepo := menuPostgres.NewMenuRepository(dbPool)
		menuService := menuApp.NewMenuService(menuRepo, auditRepo)
		menuHandler := menuHttp.NewMenuHandler(menuService, authService)
		registrars = append(registrars, menuHandler)

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
		settingsService.SetJobEnqueuer(jobRepo)
		menuService.SetJobEnqueuer(jobRepo)

		generator := sitegeneration.NewGenerator(dbPool, cfg.PublicBaseURL)
		ssgWorker := sitegeneration.NewWorker(jobRepo, generator, releaseMgr, cfg.SSGWorkerPollInterval, cfg.SSGJobMaxAttempts, logger)
		ssgWorker.SetScheduledPromoter(contentService.PromoteDueScheduledTranslations)

		// Run background SSG worker lifecycle
		go ssgWorker.Start(ctx)

		// Trigger initial build if no release exists yet
		if _, err := releaseMgr.GetActiveReleaseDir(); err != nil {
			logger.Info("no active SSG release found on startup; scheduling initial site generation")
			_, _ = jobRepo.Enqueue(ctx, nil, "site.initial_build", map[string]any{"trigger": "startup"})
		}

		// 10. Public Search
		searchRepo := searchPostgres.NewSearchRepository(dbPool)
		searchService := searchApp.NewSearchService(searchRepo)
		searchHandler := searchHttp.NewSearchHandler(searchService)
		registrars = append(registrars, searchHandler)

		// 11. Editorial Curation
		curationRepo := curationPostgres.NewCurationRepository(dbPool)
		curationService := curationApp.NewCurationService(curationRepo, auditRepo, jobRepo)
		curationHandler := curationHttp.NewCurationHandler(curationService, authService)
		registrars = append(registrars, curationHandler)

		// 12. Permanent URL Redirects (301)
		redirectRepo := redirectPostgres.NewRedirectRepository(dbPool)
		redirectHandler := redirectHttp.NewRedirectHandler(redirectRepo, authService)
		registrars = append(registrars, redirectHandler)
		contentService.SetRedirectRecorder(redirectRepo)
		pageService.SetRedirectRecorder(redirectRepo)

		// 13. System Audit Logs
		auditHandler := auditHttp.NewAuditHandler(auditRepo, authService)
		registrars = append(registrars, auditHandler)

		// 14. Document Downloads (Unduhan)
		downloadRepo := downloadPostgres.NewDownloadRepository(dbPool)
		downloadService := downloadApp.NewDownloadService(downloadRepo, auditRepo)
		downloadService.SetJobEnqueuer(jobRepo)
		downloadHandler := downloadHttp.NewDownloadHandler(downloadService, authService, localStorage)
		registrars = append(registrars, downloadHandler)

		// 15. Analytics, Traffic Tracking & Real-Time Dashboard
		analyticsRepo := analyticsPostgres.NewAnalyticsRepository(dbPool)
		analyticsService := analyticsApp.NewAnalyticsService(analyticsRepo)
		analyticsHandler := analyticsHttp.NewAnalyticsHandler(analyticsService, authService)
		registrars = append(registrars, analyticsHandler)

		// 16. Layout, Promotional Banners & AdSense Slots
		layoutRepo := layoutPostgres.NewBannerRepository(dbPool)
		layoutService := layoutApp.NewLayoutService(layoutRepo, auditRepo)
		layoutService.SetJobEnqueuer(jobRepo)
		layoutHandler := layoutHttp.NewLayoutHandler(layoutService, authService)
		registrars = append(registrars, layoutHandler)
	} else {
		logger.Warn("PIDIEKAB_DATABASE_URL is not configured; running without database connection pool")
	}

	server := httpserver.New(cfg, logger, dbPool, registrars...)

	var redirectServer *http.Server
	errCh := make(chan error, 2)

	go func() {
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			logger.Info("starting pidiekab server with native TLS (HTTPS)",
				"addr", cfg.Address,
				"cert_file", cfg.TLSCertFile,
				"environment", cfg.Environment,
				"log_level", cfg.LogLevel.String(),
			)
			errCh <- server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			logger.Info("starting pidiekab server (HTTP)",
				"addr", cfg.Address,
				"environment", cfg.Environment,
				"log_level", cfg.LogLevel.String(),
			)
			errCh <- server.ListenAndServe()
		}
	}()

	if cfg.TLSCertFile != "" && cfg.HTTPRedirectAddr != "" {
		redirectServer = httpserver.NewHTTPToHTTPSRedirectServer(cfg.HTTPRedirectAddr, cfg, logger)
		go func() {
			logger.Info("starting HTTP-to-HTTPS redirect server", "addr", cfg.HTTPRedirectAddr)
			if err := redirectServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Warn("HTTP redirect server stopped", "error", err)
			}
		}()
	}

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

		if redirectServer != nil {
			_ = redirectServer.Shutdown(shutdownCtx)
		}

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
		fmt.Println("Subcommands: create, list, reset-password")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

		admin, err := authService.BootstrapAdmin(ctx, *emailFlag, *nameFlag, *passwordFlag)
		if err != nil {
			logger.Error("failed to create admin", "error", err)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		logger.Info("administrator created successfully", "id", admin.ID, "email", admin.Email, "name", admin.DisplayName)
		fmt.Printf("Administrator %q (ID %d) created successfully.\n", admin.Email, admin.ID)

	case "list":
		admins, err := authService.ListAdmins(ctx)
		if err != nil {
			logger.Error("failed to list admins", "error", err)
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("%-6s | %-30s | %-25s | %-8s\n", "ID", "Email", "Display Name", "Status")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, a := range admins {
			status := "active"
			if !a.IsActive {
				status = "inactive"
			}
			fmt.Printf("%-6d | %-30s | %-25s | %-8s\n", a.ID, a.Email, a.DisplayName, status)
		}
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Printf("Total administrators: %d\n", len(admins))

	case "reset-password":
		cmd := flag.NewFlagSet("admin reset-password", flag.ExitOnError)
		emailFlag := cmd.String("email", "", "Admin email address to reset")
		passwordFlag := cmd.String("password", "", "New admin password (minimum 8 characters)")

		if err := cmd.Parse(os.Args[3:]); err != nil {
			fmt.Println("Failed to parse flags:", err)
			os.Exit(1)
		}

		if *emailFlag == "" || *passwordFlag == "" {
			fmt.Println("Error: --email and --password are required.")
			cmd.Usage()
			os.Exit(1)
		}

		if err := authService.ResetPasswordByOperator(ctx, *emailFlag, *passwordFlag); err != nil {
			logger.Error("failed to reset admin password", "error", err)
			fmt.Printf("Error resetting password: %v\n", err)
			os.Exit(1)
		}

		logger.Info("administrator password reset successfully", "email", *emailFlag)
		fmt.Printf("Password for administrator %q has been reset successfully. All active sessions revoked.\n", *emailFlag)

	default:
		fmt.Printf("Unknown admin subcommand %q. Available: create, list, reset-password\n", os.Args[2])
		os.Exit(1)
	}
}
