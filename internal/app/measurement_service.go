// Package app holds the use cases. It orchestrates the domain and the
// repository ports; it knows nothing about HTTP or SQL.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

// MeasurementService records readings and keeps sensor health in sync.
type MeasurementService struct {
	measurements domain.MeasurementRepository
	sensors      domain.SensorRepository
	alerts       domain.AlertRepository
	uow          domain.UnitOfWork
	now          func() time.Time
}

func NewMeasurementService(
	measurements domain.MeasurementRepository,
	sensors domain.SensorRepository,
	alerts domain.AlertRepository,
	uow domain.UnitOfWork,
) *MeasurementService {
	return &MeasurementService{
		measurements: measurements,
		sensors:      sensors,
		alerts:       alerts,
		uow:          uow,
		now:          time.Now,
	}
}

// Record stores a reading, re-evaluates the sensor's status and raises an
// alert when it turns unhealthy. Unknown sensors register themselves, which is
// how devices come online.
func (s *MeasurementService) Record(ctx context.Context, cmd RecordMeasurementCommand) (MeasurementDTO, error) {
	sensorID := cmd.SensorID

	m, err := domain.NewMeasurement(sensorID, cmd.Level, s.now().UTC())
	if err != nil {
		return MeasurementDTO{}, err
	}

	err = s.uow.Do(ctx, func(ctx context.Context) error {
		sensor, err := s.sensors.ByID(ctx, sensorID)
		if errors.Is(err, domain.ErrNotFound) {
			sensor = domain.NewSensor(sensorID)
			err = s.sensors.Save(ctx, sensor)
		}
		if err != nil {
			return err
		}

		if err := s.measurements.Add(ctx, &m); err != nil {
			return err
		}

		recent, err := s.measurements.Recent(ctx, sensorID, domain.AlertStreak)
		if err != nil {
			return err
		}

		status := domain.EvaluateStatus(recent)
		if status == sensor.Status {
			return nil
		}

		sensor.Status = status
		if err := s.sensors.Save(ctx, sensor); err != nil {
			return err
		}
		if status != domain.StatusAlert {
			return nil
		}

		// The status changed to ALERT, so this streak has not been reported yet.
		alert, err := domain.NewAlert(recent)
		if err != nil {
			return err
		}
		return s.alerts.Add(ctx, &alert)
	})
	if err != nil {
		return MeasurementDTO{}, err
	}
	return measurementDTO(m), nil
}

// BySensor lists every reading of a sensor, newest first.
func (s *MeasurementService) BySensor(ctx context.Context, sensorID uint) ([]MeasurementDTO, error) {
	measurements, err := s.measurements.BySensor(ctx, sensorID)
	if err != nil {
		return nil, err
	}
	return measurementDTOs(measurements), nil
}
