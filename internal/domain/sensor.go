// Package domain holds the business rules for CO2 sensors: entities, the
// invariants that guard them and the repository ports the outer layers plug
// into. It imports nothing outside the standard library on purpose.
package domain

import "errors"

// Status is the health of a sensor, derived from its latest measurements.
type Status string

const (
	StatusOK    Status = "OK"
	StatusWarn  Status = "WARN"
	StatusAlert Status = "ALERT"
)

const (
	// WarnLevel is the CO2 concentration (ppm) above which a reading is unhealthy.
	WarnLevel uint = 2000
	// AlertStreak is how many consecutive unhealthy readings raise an alert.
	AlertStreak = 3
)

// ErrNotFound is returned by repositories when an entity does not exist.
var ErrNotFound = errors.New("not found")

// Sensor is the aggregate root.
type Sensor struct {
	ID     uint
	Status Status
}

func NewSensor(id uint) Sensor {
	return Sensor{ID: id, Status: StatusOK}
}

// EvaluateStatus reports the status implied by recent, the sensor's latest
// measurements ordered newest first. The streak breaks on the first healthy
// reading, so a sensor recovers as soon as it reports a normal level.
func EvaluateStatus(recent []Measurement) Status {
	if len(recent) == 0 || !recent[0].Unhealthy() {
		return StatusOK
	}
	if len(recent) < AlertStreak {
		return StatusWarn
	}
	for _, m := range recent[:AlertStreak] {
		if !m.Unhealthy() {
			return StatusWarn
		}
	}
	return StatusAlert
}
