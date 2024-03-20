package service

import (
	"alarm-api/repository"
	"alarm-api/response"

	"github.com/go-playground/validator/v10"
)

type AlertServiceInterface interface {
	FindAllBySensorId(sensorId int) []response.AlertResponse
}
type AlertService struct {
	AlertRepository repository.AlertRepositoryInterface
	Validate        *validator.Validate
}

func New(alertRepository repository.AlertRepositoryInterface, validate *validator.Validate) AlertServiceInterface {
	return &AlertService{
		AlertRepository: alertRepository,
		Validate:        validate,
	}
}

func (a *AlertService) FindAllBySensorId(sensorId int) []response.AlertResponse {
	result := a.AlertRepository.FindAllBySensorId(sensorId)

	var alerts []response.AlertResponse
	for _, value := range result {
		alert := response.AlertResponse{
			Id:           value.Id,
			SensorId:     value.Sensor.Id,
			Measurement1: value.Measurement1,
			Measurement2: value.Measurement2,
			Measurement3: value.Measurement3,
			StartTime:    value.StartTime,
			EndTime:      value.EndTime,
		}
		alerts = append(alerts, alert)
	}

	return alerts
}
