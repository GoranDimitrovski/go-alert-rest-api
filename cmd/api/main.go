// Command api serves the sensor alert REST API.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	httpadapter "github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/http"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/postgres"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}

// run owns the wiring: every dependency is built here and injected downwards,
// so no package has to reach out for one.
func run() error {
	cfg := config.Load()
	setupLogging(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Info().Str("dsn", cfg.Redacted()).Msg("connecting to database")
	db, err := postgres.Open(dialCtx, cfg.DSN)
	if err != nil {
		return err
	}
	if err := postgres.Migrate(db); err != nil {
		return err
	}

	sensors := postgres.NewSensorRepository(db)
	measurements := postgres.NewMeasurementRepository(db)
	alerts := postgres.NewAlertRepository(db)
	uow := postgres.NewUnitOfWork(db)

	router := httpadapter.NewRouter(
		app.NewMeasurementService(measurements, sensors, alerts, uow),
		app.NewAlertService(alerts),
		app.NewSensorService(sensors),
		postgres.Ping(db),
	)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", server.Addr).Msg("listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info().Msg("shutting down")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return postgres.Close(db)
}

func setupLogging(level string) {
	parsed, err := zerolog.ParseLevel(level)
	if err != nil {
		parsed = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(parsed)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	gin.SetMode(gin.ReleaseMode)
}
