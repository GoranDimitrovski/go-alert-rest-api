// Package http is the delivery adapter: it turns HTTP requests into use-case
// calls and domain errors into status codes. No business rule lives here.
package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

// NewRouter wires the routes. ping reports whether dependencies are reachable.
func NewRouter(
	measurements *app.MeasurementService,
	alerts *app.AlertService,
	sensors *app.SensorService,
	ping func(context.Context) error,
) *gin.Engine {
	h := &handler{measurements: measurements, alerts: alerts, sensors: sensors}

	router := gin.New()
	router.Use(requestLogger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		if err := ping(c.Request.Context()); err != nil {
			log.Error().Err(err).Msg("health check failed")
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "database unavailable"})
			return
		}
		c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/measurements", h.recordMeasurement)
		v1.GET("/sensors/:sensor_id", h.getSensor)
		v1.GET("/sensors/:sensor_id/measurements", h.listMeasurements)
		v1.GET("/sensors/:sensor_id/alerts", h.listAlerts)
	}

	return router
}

// requestLogger keeps request logging in the same structured format as the
// rest of the service.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		event := log.Info()
		if c.Writer.Status() >= http.StatusInternalServerError {
			event = log.Error()
		}
		event.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start)).
			Msg("request")
	}
}

// statusClientClosedRequest is the non-standard code for a client that hung up
// before the response was written.
const statusClientClosedRequest = 499

// fail maps a domain error onto the response. Unknown errors stay opaque to
// the client and detailed in the log.
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
	case errors.Is(err, domain.ErrInvalidSensor), errors.Is(err, domain.ErrInvalidLevel):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, context.Canceled):
		c.Status(statusClientClosedRequest) // client gone; nothing to report
	default:
		log.Error().Err(err).Str("path", c.Request.URL.Path).Msg("unhandled error")
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
