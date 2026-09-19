// Package config loads settings from the environment. Nothing else in the
// service reads os.Getenv.
package config

import (
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Addr            string
	DSN             string
	LogLevel        string
	ShutdownTimeout time.Duration
}

// Load reads the environment, falling back to values that work against the
// docker compose stack.
func Load() Config {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(env("DB_USER", "myuser"), env("DB_PASSWORD", "mypassword")),
		Host:     net.JoinHostPort(env("DB_HOST", "postgres"), env("DB_PORT", "5432")),
		Path:     env("DB_NAME", "alarm"),
		RawQuery: "sslmode=" + env("DB_SSLMODE", "disable"),
	}

	return Config{
		Addr:            net.JoinHostPort("", env("PORT", "8080")),
		DSN:             dsn.String(),
		LogLevel:        env("LOG_LEVEL", "info"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

// Redacted returns the DSN with the password masked, safe to log.
func (c Config) Redacted() string {
	u, err := url.Parse(c.DSN)
	if err != nil {
		return ""
	}
	return u.Redacted()
}
