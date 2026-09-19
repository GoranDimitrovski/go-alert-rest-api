package http

import (
	"time"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

// These types are the wire contract. They are exported so tests and clients can
// decode into them instead of guessing at field names.

type MeasurementRequest struct {
	SensorID uint `json:"sensor_id" binding:"required"`
	Level    uint `json:"level"     binding:"required"`
}

func (r MeasurementRequest) Command() app.RecordMeasurementCommand {
	return app.RecordMeasurementCommand{SensorID: r.SensorID, Level: r.Level}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type SensorResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type MeasurementResponse struct {
	ID         uint      `json:"id"`
	SensorID   uint      `json:"sensor_id"`
	Level      uint      `json:"level"`
	RecordedAt time.Time `json:"recorded_at"`
}

// AlertResponse carries its levels oldest first.
type AlertResponse struct {
	ID        uint                     `json:"id"`
	SensorID  uint                     `json:"sensor_id"`
	Levels    [domain.AlertStreak]uint `json:"levels"`
	StartTime time.Time                `json:"start_time"`
	EndTime   time.Time                `json:"end_time"`
}

func newSensorResponse(s app.SensorDTO) SensorResponse {
	return SensorResponse{ID: s.ID, Status: s.Status}
}

func newMeasurementResponse(m app.MeasurementDTO) MeasurementResponse {
	return MeasurementResponse{ID: m.ID, SensorID: m.SensorID, Level: m.Level, RecordedAt: m.RecordedAt}
}

func newMeasurementResponses(measurements []app.MeasurementDTO) []MeasurementResponse {
	out := make([]MeasurementResponse, 0, len(measurements))
	for _, m := range measurements {
		out = append(out, newMeasurementResponse(m))
	}
	return out
}

func newAlertResponses(alerts []app.AlertDTO) []AlertResponse {
	out := make([]AlertResponse, 0, len(alerts))
	for _, a := range alerts {
		out = append(out, AlertResponse{
			ID:        a.ID,
			SensorID:  a.SensorID,
			Levels:    a.Levels,
			StartTime: a.StartTime,
			EndTime:   a.EndTime,
		})
	}
	return out
}
