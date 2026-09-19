package domain

import (
	"errors"
	"time"
)

// MaxLevel rejects readings no sensor can legitimately produce.
const MaxLevel uint = 100_000

var (
	ErrInvalidSensor = errors.New("sensor id must be greater than zero")
	ErrInvalidLevel  = errors.New("level must be between 1 and 100000 ppm")
)

type Measurement struct {
	ID         uint
	SensorID   uint
	Level      uint
	RecordedAt time.Time
}

func NewMeasurement(sensorID, level uint, at time.Time) (Measurement, error) {
	if sensorID == 0 {
		return Measurement{}, ErrInvalidSensor
	}
	if level == 0 || level > MaxLevel {
		return Measurement{}, ErrInvalidLevel
	}
	return Measurement{SensorID: sensorID, Level: level, RecordedAt: at}, nil
}

func (m Measurement) Unhealthy() bool { return m.Level > WarnLevel }
