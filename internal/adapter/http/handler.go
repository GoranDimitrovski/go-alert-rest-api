package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type handler struct {
	measurements *app.MeasurementService
	alerts       *app.AlertService
	sensors      *app.SensorService
}

func (h *handler) recordMeasurement(c *gin.Context) {
	var req MeasurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// The validator's message names internal fields; keep it in the log.
		log.Debug().Err(err).Msg("rejected measurement payload")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "body must be {\"sensor_id\": <uint>, \"level\": <uint>}"})
		return
	}

	measurement, err := h.measurements.Record(c.Request.Context(), req.Command())
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, newMeasurementResponse(measurement))
}

func (h *handler) listMeasurements(c *gin.Context) {
	sensorID, ok := sensorIDParam(c)
	if !ok {
		return
	}

	measurements, err := h.measurements.BySensor(c.Request.Context(), sensorID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, newMeasurementResponses(measurements))
}

func (h *handler) listAlerts(c *gin.Context) {
	sensorID, ok := sensorIDParam(c)
	if !ok {
		return
	}

	alerts, err := h.alerts.BySensor(c.Request.Context(), sensorID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, newAlertResponses(alerts))
}

func (h *handler) getSensor(c *gin.Context) {
	sensorID, ok := sensorIDParam(c)
	if !ok {
		return
	}

	sensor, err := h.sensors.ByID(c.Request.Context(), sensorID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, newSensorResponse(sensor))
}

// sensorIDParam parses the path parameter and answers the request itself when
// it is malformed.
func sensorIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("sensor_id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: domain.ErrInvalidSensor.Error()})
		return 0, false
	}
	return uint(id), true
}
