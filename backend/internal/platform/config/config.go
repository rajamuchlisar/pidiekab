package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all platform and runtime configurations for the PidieKab service.
type Config struct {
	Environment               string
	Address                   string
	LogLevel                  slog.Level
	ReadHeaderTimeout         time.Duration
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	IdleTimeout               time.Duration
	ShutdownTimeout           time.Duration
	DatabaseURL               string
	DatabaseMaxConns          int32
	DatabaseMinConns          int32
	DatabaseMaxConnLifetime   time.Duration
	DatabaseMaxConnIdleTime   time.Duration
	DatabaseHealthCheckPeriod time.Duration
	PublicStaticDir           string
	CMSStaticDir              string
}

// Load reads configuration from environment variables, applies validated defaults,
// and enforces strict validation. Startup fails explicitly if required values are invalid.
func Load() (Config, error) {
	envName := strings.ToLower(env("PIDIEKAB_ENV", "development"))
	switch envName {
	case "development", "test", "production":
		// valid
	default:
		return Config{}, fmt.Errorf("PIDIEKAB_ENV must be one of: development, test, production (got: %q)", envName)
	}

	addr := env("PIDIEKAB_ADDR", "127.0.0.1:8080")
	if strings.TrimSpace(addr) == "" {
		return Config{}, fmt.Errorf("PIDIEKAB_ADDR must not be empty")
	}

	logLevelStr := strings.ToLower(env("PIDIEKAB_LOG_LEVEL", "info"))
	var logLevel slog.Level
	switch logLevelStr {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return Config{}, fmt.Errorf("PIDIEKAB_LOG_LEVEL must be debug, info, warn, or error (got: %q)", logLevelStr)
	}

	readHeaderTimeout, err := parseDurationEnv("PIDIEKAB_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}
	if readHeaderTimeout <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_READ_HEADER_TIMEOUT must be positive (got: %v)", readHeaderTimeout)
	}

	readTimeout, err := parseDurationEnv("PIDIEKAB_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	if readTimeout <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_READ_TIMEOUT must be positive (got: %v)", readTimeout)
	}

	writeTimeout, err := parseDurationEnv("PIDIEKAB_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return Config{}, err
	}
	if writeTimeout <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_WRITE_TIMEOUT must be positive (got: %v)", writeTimeout)
	}

	idleTimeout, err := parseDurationEnv("PIDIEKAB_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}
	if idleTimeout <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_IDLE_TIMEOUT must be positive (got: %v)", idleTimeout)
	}

	shutdownTimeout, err := parseDurationEnv("PIDIEKAB_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	if shutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_SHUTDOWN_TIMEOUT must be positive (got: %v)", shutdownTimeout)
	}

	dbURL := env("PIDIEKAB_DATABASE_URL", "")

	maxConns, err := parseIntEnv("PIDIEKAB_DATABASE_MAX_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	if maxConns <= 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_DATABASE_MAX_CONNS must be greater than 0")
	}

	minConns, err := parseIntEnv("PIDIEKAB_DATABASE_MIN_CONNS", 2)
	if err != nil {
		return Config{}, err
	}
	if minConns < 0 {
		return Config{}, fmt.Errorf("PIDIEKAB_DATABASE_MIN_CONNS must be non-negative")
	}
	if minConns > maxConns {
		return Config{}, fmt.Errorf("PIDIEKAB_DATABASE_MIN_CONNS (%d) cannot exceed PIDIEKAB_DATABASE_MAX_CONNS (%d)", minConns, maxConns)
	}

	maxConnLifetime, err := parseDurationEnv("PIDIEKAB_DATABASE_MAX_CONN_LIFETIME", 1*time.Hour)
	if err != nil {
		return Config{}, err
	}

	maxConnIdleTime, err := parseDurationEnv("PIDIEKAB_DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	healthCheckPeriod, err := parseDurationEnv("PIDIEKAB_DATABASE_HEALTH_CHECK_PERIOD", 1*time.Minute)
	if err != nil {
		return Config{}, err
	}

	publicStaticDir := env("PIDIEKAB_PUBLIC_STATIC_DIR", "frontend/public-site/dist")
	cmsStaticDir := env("PIDIEKAB_CMS_STATIC_DIR", "frontend/admin-cms/dist")

	return Config{
		Environment:               envName,
		Address:                   addr,
		LogLevel:                  logLevel,
		ReadHeaderTimeout:         readHeaderTimeout,
		ReadTimeout:               readTimeout,
		WriteTimeout:              writeTimeout,
		IdleTimeout:               idleTimeout,
		ShutdownTimeout:           shutdownTimeout,
		DatabaseURL:               dbURL,
		DatabaseMaxConns:          int32(maxConns),
		DatabaseMinConns:          int32(minConns),
		DatabaseMaxConnLifetime:   maxConnLifetime,
		DatabaseMaxConnIdleTime:   maxConnIdleTime,
		DatabaseHealthCheckPeriod: healthCheckPeriod,
		PublicStaticDir:           publicStaticDir,
		CMSStaticDir:              cmsStaticDir,
	}, nil
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func parseIntEnv(key string, fallback int) (int, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s (%q): %w", key, val, err)
	}
	return i, nil
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s (%q): %w", key, val, err)
	}
	return d, nil
}

// ParseBoolEnv is reserved for feature flags added in later phases.
func ParseBoolEnv(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}
