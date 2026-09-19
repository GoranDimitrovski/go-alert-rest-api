package app

import (
	"time"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

// The use cases take commands and hand back DTOs, so domain entities never
// leave this layer and the adapters cannot reach past the API they are given.

type RecordMeasurementCommand struct {
	SensorID uint
	Level    uint
}

type MeasurementDTO struct {
	ID         uint
	SensorID   uint
	Level      uint
	RecordedAt time.Time
}

type SensorDTO struct {
	ID     uint
	Status string
}

// AlertDTO carries its levels oldest first.
type AlertDTO struct {
	ID        uint
	SensorID  uint
	Levels    [domain.AlertStreak]uint
	StartTime time.Time
	EndTime   time.Time
}

func measurementDTO(m domain.Measurement) MeasurementDTO {
	return MeasurementDTO{ID: m.ID, SensorID: m.SensorID, Level: m.Level, RecordedAt: m.RecordedAt}
}

func measurementDTOs(measurements []domain.Measurement) []MeasurementDTO {
	out := make([]MeasurementDTO, len(measurements))
	for i, m := range measurements {
		out[i] = measurementDTO(m)
	}
	return out
}

func sensorDTO(s domain.Sensor) SensorDTO {
	return SensorDTO{ID: s.ID, Status: string(s.Status)}
}

func alertDTOs(alerts []domain.Alert) []AlertDTO {
	out := make([]AlertDTO, len(alerts))
	for i, a := range alerts {
		out[i] = AlertDTO{
			ID:        a.ID,
			SensorID:  a.SensorID,
			Levels:    a.Levels,
			StartTime: a.StartTime,
			EndTime:   a.EndTime,
		}
	}
	return out
}
