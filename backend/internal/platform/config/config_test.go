package config_test

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"pidiekab/internal/platform/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Clean environment
	for _, k := range []string{
		"PIDIEKAB_ENV", "PIDIEKAB_ADDR", "PIDIEKAB_LOG_LEVEL",
		"PIDIEKAB_READ_HEADER_TIMEOUT", "PIDIEKAB_READ_TIMEOUT",
		"PIDIEKAB_WRITE_TIMEOUT", "PIDIEKAB_IDLE_TIMEOUT",
		"PIDIEKAB_SHUTDOWN_TIMEOUT", "PIDIEKAB_DATABASE_URL",
	} {
		os.Unsetenv(k)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected valid default config, got err: %v", err)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected environment 'development', got %q", cfg.Environment)
	}
	if cfg.Address != "127.0.0.1:8080" {
		t.Errorf("expected address '127.0.0.1:8080', got %q", cfg.Address)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("expected log level info, got %v", cfg.LogLevel)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("expected read header timeout 5s, got %v", cfg.ReadHeaderTimeout)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("expected read timeout 15s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 30*time.Second {
		t.Errorf("expected write timeout 30s, got %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("expected idle timeout 60s, got %v", cfg.IdleTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected shutdown timeout 10s, got %v", cfg.ShutdownTimeout)
	}
}

func TestConfigValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		envKey  string
		envVal  string
		wantErr bool
	}{
		{"invalid environment", "PIDIEKAB_ENV", "staging", true},
		{"empty address", "PIDIEKAB_ADDR", "   ", true},
		{"invalid log level", "PIDIEKAB_LOG_LEVEL", "trace", true},
		{"invalid duration format", "PIDIEKAB_READ_TIMEOUT", "invalid", true},
		{"negative duration", "PIDIEKAB_WRITE_TIMEOUT", "-5s", true},
		{"zero shutdown timeout", "PIDIEKAB_SHUTDOWN_TIMEOUT", "0s", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			os.Setenv(tc.envKey, tc.envVal)
			_, err := config.Load()
			if tc.wantErr && err == nil {
				t.Errorf("expected error for %s=%s, got nil", tc.envKey, tc.envVal)
			}
		})
	}
}

func TestConfigCustomValidValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("PIDIEKAB_ENV", "production")
	os.Setenv("PIDIEKAB_ADDR", ":9090")
	os.Setenv("PIDIEKAB_LOG_LEVEL", "debug")
	os.Setenv("PIDIEKAB_READ_HEADER_TIMEOUT", "3s")
	os.Setenv("PIDIEKAB_READ_TIMEOUT", "10s")
	os.Setenv("PIDIEKAB_WRITE_TIMEOUT", "20s")
	os.Setenv("PIDIEKAB_IDLE_TIMEOUT", "45s")
	os.Setenv("PIDIEKAB_SHUTDOWN_TIMEOUT", "15s")
	os.Setenv("PIDIEKAB_DATABASE_URL", "postgres://localhost:5432/pidie")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected production, got %q", cfg.Environment)
	}
	if cfg.Address != ":9090" {
		t.Errorf("expected :9090, got %q", cfg.Address)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("expected debug log level, got %v", cfg.LogLevel)
	}
	if cfg.ReadHeaderTimeout != 3*time.Second {
		t.Errorf("expected 3s, got %v", cfg.ReadHeaderTimeout)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/pidie" {
		t.Errorf("expected db url, got %q", cfg.DatabaseURL)
	}
}
