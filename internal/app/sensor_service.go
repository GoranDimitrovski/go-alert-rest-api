package app

import (
	"context"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type SensorService struct {
	sensors domain.SensorRepository
}

func NewSensorService(sensors domain.SensorRepository) *SensorService {
	return &SensorService{sensors: sensors}
}

// ByID returns domain.ErrNotFound for an unknown sensor.
func (s *SensorService) ByID(ctx context.Context, id uint) (SensorDTO, error) {
	sensor, err := s.sensors.ByID(ctx, id)
	if err != nil {
		return SensorDTO{}, err
	}
	return sensorDTO(sensor), nil
}
