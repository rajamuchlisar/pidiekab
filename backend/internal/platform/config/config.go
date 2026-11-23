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
	PublicBaseURL             string
	StoragePath               string
	SSGOutputRoot             string
	SSGReleasesToKeep         int
	SSGWorkerPollInterval     time.Duration
	SSGJobMaxAttempts         int
	TLSCertFile               string
	TLSKeyFile                string
	HTTPRedirectAddr          string
	AllowedHosts              []string
}

// Load reads configuration from environment variables, applies validated defaults,
// and enforces strict validation. Startup fails explicitly if required values are invalid.
func Load() (Config, error) {
	loadDotEnvIfExists()

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
	publicBaseURL := strings.TrimRight(env("PIDIEKAB_BASE_URL", env("PIDIEKAB_PUBLIC_BASE_URL", "https://pidiekab.web.id")), "/")
	storagePath := env("PIDIEKAB_STORAGE_PATH", "data/media")
	ssgOutputRoot := env("PIDIEKAB_SSG_OUTPUT_ROOT", "data/ssg")

	ssgReleasesToKeep, err := parseIntEnv("PIDIEKAB_SSG_RELEASES_TO_KEEP", 5)
	if err != nil || ssgReleasesToKeep < 1 {
		ssgReleasesToKeep = 5
	}

	ssgWorkerPollInterval, err := parseDurationEnv("PIDIEKAB_SSG_POLL_INTERVAL", 5*time.Second)
	if err != nil || ssgWorkerPollInterval <= 0 {
		ssgWorkerPollInterval = 5 * time.Second
	}

	ssgJobMaxAttempts, err := parseIntEnv("PIDIEKAB_SSG_JOB_MAX_ATTEMPTS", 3)
	if err != nil || ssgJobMaxAttempts < 1 {
		ssgJobMaxAttempts = 3
	}

	tlsCertFile := env("PIDIEKAB_TLS_CERT_FILE", "")
	tlsKeyFile := env("PIDIEKAB_TLS_KEY_FILE", "")
	if (tlsCertFile != "" && tlsKeyFile == "") || (tlsCertFile == "" && tlsKeyFile != "") {
		return Config{}, fmt.Errorf("both PIDIEKAB_TLS_CERT_FILE and PIDIEKAB_TLS_KEY_FILE must be set together")
	}

	httpRedirectAddr := env("PIDIEKAB_HTTP_REDIRECT_ADDR", "")
	if tlsCertFile != "" && httpRedirectAddr == "" {
		httpRedirectAddr = ":80"
	}

	allowedHostsStr := env("PIDIEKAB_ALLOWED_HOSTS", "")
	var allowedHosts []string
	if allowedHostsStr != "" {
		for _, h := range strings.Split(allowedHostsStr, ",") {
			h = strings.ToLower(strings.TrimSpace(h))
			if h != "" {
				allowedHosts = append(allowedHosts, h)
			}
		}
	} else if envName == "production" {
		allowedHosts = []string{"pidiekab.web.id", "www.pidiekab.web.id", "localhost", "127.0.0.1", "::1"}
	}

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
		PublicBaseURL:             publicBaseURL,
		StoragePath:               storagePath,
		SSGOutputRoot:             ssgOutputRoot,
		SSGReleasesToKeep:         ssgReleasesToKeep,
		SSGWorkerPollInterval:     ssgWorkerPollInterval,
		SSGJobMaxAttempts:         ssgJobMaxAttempts,
		TLSCertFile:               tlsCertFile,
		TLSKeyFile:                tlsKeyFile,
		HTTPRedirectAddr:          httpRedirectAddr,
		AllowedHosts:              allowedHosts,
	}, nil
}

// IsHostAllowed reports whether the given hostname is permitted to access the application.
// If AllowedHosts is empty, all hosts are allowed (convenient for local unit tests).
func (c Config) IsHostAllowed(host string) bool {
	if len(c.AllowedHosts) == 0 {
		return true
	}
	if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
		host = host[:colonIdx]
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	for _, allowed := range c.AllowedHosts {
		if allowed == "*" || allowed == host {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[1:]
			if strings.HasSuffix(host, suffix) {
				return true
			}
		}
	}
	return false
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

// loadDotEnvIfExists reads a .env file from the current directory or parent directory if present,
// populating environment variables that have not already been explicitly defined.
func loadDotEnvIfExists() {
	candidates := []string{".env", "../.env"}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, rawLine := range lines {
			line := strings.TrimSpace(rawLine)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)
				if _, set := os.LookupEnv(k); !set {
					_ = os.Setenv(k, v)
				}
			}
		}
		break
	}
}
