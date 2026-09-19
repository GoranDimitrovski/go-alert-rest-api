package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg := config.Load()

	if cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":8080")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
	if !strings.HasPrefix(cfg.DSN, "postgres://") {
		t.Errorf("DSN = %q, want a postgres URL", cfg.DSN)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "alice")
	t.Setenv("DB_PASSWORD", "p@ss word/!")
	t.Setenv("DB_NAME", "sensors")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("SHUTDOWN_TIMEOUT", "45s")

	cfg := config.Load()

	if cfg.Addr != ":9090" {
		t.Errorf("Addr = %q, want %q", cfg.Addr, ":9090")
	}
	if cfg.ShutdownTimeout != 45*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 45s", cfg.ShutdownTimeout)
	}

	// A password with URL-hostile characters has to survive intact.
	want := "postgres://alice:p%40ss%20word%2F%21@db.internal:6543/sensors?sslmode=require"
	if cfg.DSN != want {
		t.Errorf("DSN  = %q\nwant = %q", cfg.DSN, want)
	}
}

func TestEmptyEnvironmentVariableFallsBack(t *testing.T) {
	t.Setenv("PORT", "")

	if cfg := config.Load(); cfg.Addr != ":8080" {
		t.Errorf("Addr = %q, want the default when PORT is empty", cfg.Addr)
	}
}

func TestInvalidDurationFallsBack(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT", "ten seconds")

	if cfg := config.Load(); cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want the 10s default", cfg.ShutdownTimeout)
	}
}

func TestRedactedHidesThePassword(t *testing.T) {
	t.Setenv("DB_PASSWORD", "hunter2")

	redacted := config.Load().Redacted()
	if strings.Contains(redacted, "hunter2") {
		t.Fatalf("Redacted() leaked the password: %q", redacted)
	}
	if !strings.Contains(redacted, "xxxxx") {
		t.Errorf("Redacted() = %q, want a masked password", redacted)
	}
}
