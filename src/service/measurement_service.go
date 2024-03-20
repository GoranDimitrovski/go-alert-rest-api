package service

import (
	"alarm-api/helper"
	"alarm-api/model"
	"alarm-api/repository"
	"alarm-api/request"
	"alarm-api/response"

	"github.com/go-playground/validator/v10"
)

type MeasurementServiceInterface interface {
	Create(measurement request.CreateMeasurementRequest)
	FindAll(sensorId int) []response.MeasurementResponse
}

type MeasurementService struct {
	MeasurementRepository repository.MeasurmentRepositoryInterface
	SensorRepository      repository.SensorRepositoryInterface
	Validate              *validator.Validate
}

func NewMeasurmentService(
	measurementRepository repository.MeasurmentRepositoryInterface,
	sensorRepository repository.SensorRepositoryInterface,
	validate *validator.Validate,
) MeasurementServiceInterface {
	return &MeasurementService{
		MeasurementRepository: measurementRepository,
		SensorRepository:      sensorRepository,
		Validate:              validate,
	}
}

func (m *MeasurementService) Create(measurementRequest request.CreateMeasurementRequest) {
	err := m.Validate.Struct(measurementRequest)
	helper.ErrorPanic(err)
	sensor, err := m.SensorRepository.FindBySensorId(int(measurementRequest.SensorId))

	if err != nil {
		status := model.Status{
			Name: "OK",
		}

		sensor = model.Sensor{
			Id:     measurementRequest.SensorId,
			Status: status,
		}

		m.SensorRepository.Save(sensor)
	}

	measurement := model.Measurement{
		Level:    measurementRequest.Level,
		SensorID: sensor.Id,
		Sensor:   sensor,
	}

	if measurement.Level > 2000 {
		sensor.StatusId = 2
	}

	m.MeasurementRepository.Save(measurement)

	if m.MeasurementRepository.IsAlert(int(sensor.Id)) {
		m.SensorRepository.SetAlert(sensor)
	}
}

func (m *MeasurementService) FindAll(sensorId int) []response.MeasurementResponse {
	result := m.MeasurementRepository.FindAll(sensorId)

	var measurements []response.MeasurementResponse
	for _, value := range result {
		measurement := response.MeasurementResponse{
			Id:       value.Id,
			Level:    value.Level,
			SensorId: value.SensorID,
		}
		measurements = append(measurements, measurement)
	}

	return measurements
}
