// Package memory implements the domain repository ports in memory. It exists
// so the use cases and the HTTP layer can be tested without a database.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

var (
	_ domain.SensorRepository      = SensorRepository{}
	_ domain.MeasurementRepository = MeasurementRepository{}
	_ domain.AlertRepository       = AlertRepository{}
	_ domain.UnitOfWork            = UnitOfWork{}
)

type Store struct {
	mu           sync.Mutex
	sensors      map[uint]domain.Sensor
	measurements []domain.Measurement
	alerts       []domain.Alert
	nextID       uint

	// FailWith, when set, is returned by every write. It is how tests reach
	// the error paths of the layers above.
	FailWith error
}

func New() *Store {
	return &Store{sensors: map[uint]domain.Sensor{}}
}

func (s *Store) Sensors() SensorRepository           { return SensorRepository{s} }
func (s *Store) Measurements() MeasurementRepository { return MeasurementRepository{s} }
func (s *Store) Alerts() AlertRepository             { return AlertRepository{s} }

// UnitOfWork runs the work directly: there is nothing to roll back.
func (s *Store) UnitOfWork() UnitOfWork { return UnitOfWork{} }

type UnitOfWork struct{}

func (UnitOfWork) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type SensorRepository struct{ store *Store }

func (r SensorRepository) ByID(_ context.Context, id uint) (domain.Sensor, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	sensor, ok := r.store.sensors[id]
	if !ok {
		return domain.Sensor{}, domain.ErrNotFound
	}
	return sensor, nil
}

func (r SensorRepository) Save(_ context.Context, sensor domain.Sensor) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.FailWith != nil {
		return r.store.FailWith
	}
	r.store.sensors[sensor.ID] = sensor
	return nil
}

type MeasurementRepository struct{ store *Store }

func (r MeasurementRepository) Add(_ context.Context, m *domain.Measurement) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.FailWith != nil {
		return r.store.FailWith
	}
	r.store.nextID++
	m.ID = r.store.nextID
	r.store.measurements = append(r.store.measurements, *m)
	return nil
}

func (r MeasurementRepository) Recent(ctx context.Context, sensorID uint, limit int) ([]domain.Measurement, error) {
	all, err := r.BySensor(ctx, sensorID)
	if err != nil || len(all) <= limit {
		return all, err
	}
	return all[:limit], nil
}

// BySensor returns the newest measurements first, like the SQL repository.
func (r MeasurementRepository) BySensor(_ context.Context, sensorID uint) ([]domain.Measurement, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	var out []domain.Measurement
	for _, m := range r.store.measurements {
		if m.SensorID == sensorID {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

type AlertRepository struct{ store *Store }

func (r AlertRepository) Add(_ context.Context, a *domain.Alert) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if r.store.FailWith != nil {
		return r.store.FailWith
	}
	r.store.nextID++
	a.ID = r.store.nextID
	r.store.alerts = append(r.store.alerts, *a)
	return nil
}

func (r AlertRepository) BySensor(_ context.Context, sensorID uint) ([]domain.Alert, error) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	var out []domain.Alert
	for _, a := range r.store.alerts {
		if a.SensorID == sensorID {
			out = append(out, a)
		}
	}
	return out, nil
}
