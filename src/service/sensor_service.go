package service

import (
	"alarm-api/repository"
	"alarm-api/response"

	"github.com/go-playground/validator/v10"
)

type SensorServiceInterface interface {
	FindBySensorId(sensorId int) (response.SensorResponse, error)
}

type SensorService struct {
	SensorRepository repository.SensorRepositoryInterface
	Validate         *validator.Validate
}

func NewSensorService(
	sensorRepository repository.SensorRepositoryInterface,
	validate *validator.Validate,
) SensorServiceInterface {
	return &SensorService{
		SensorRepository: sensorRepository,
		Validate:         validate,
	}
}

func (ss *SensorService) FindBySensorId(sensorId int) (response.SensorResponse, error) {
	sensor, err := ss.SensorRepository.FindBySensorId(sensorId)

	return response.SensorResponse{
		Id:       sensor.Id,
		StatusId: sensor.Status.Id,
	}, err
}