package app

import (
	"context"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type AlertService struct {
	alerts domain.AlertRepository
}

func NewAlertService(alerts domain.AlertRepository) *AlertService {
	return &AlertService{alerts: alerts}
}

// BySensor lists the alerts raised for a sensor, newest first.
func (s *AlertService) BySensor(ctx context.Context, sensorID uint) ([]AlertDTO, error) {
	alerts, err := s.alerts.BySensor(ctx, sensorID)
	if err != nil {
		return nil, err
	}
	return alertDTOs(alerts), nil
}
