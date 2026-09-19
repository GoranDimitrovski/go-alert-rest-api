package domain

import "context"

type SensorRepository interface {
	// ByID returns ErrNotFound when the sensor is unknown.
	ByID(ctx context.Context, id uint) (Sensor, error)
	// Save inserts the sensor or updates the status of an existing one.
	Save(ctx context.Context, sensor Sensor) error
}

type MeasurementRepository interface {
	// Add persists m and fills in its ID.
	Add(ctx context.Context, m *Measurement) error
	// Recent returns at most limit measurements, newest first.
	Recent(ctx context.Context, sensorID uint, limit int) ([]Measurement, error)
	BySensor(ctx context.Context, sensorID uint) ([]Measurement, error)
}

type AlertRepository interface {
	Add(ctx context.Context, a *Alert) error
	BySensor(ctx context.Context, sensorID uint) ([]Alert, error)
}

// UnitOfWork runs fn atomically. Repositories called with the context fn
// receives join the same transaction.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
