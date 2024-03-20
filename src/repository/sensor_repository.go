package repository

import (
	"alarm-api/helper"
	"alarm-api/model"
	"errors"

	"gorm.io/gorm"
)

type SensorRepositoryInterface interface {
	FindBySensorId(sensorId int) (model.Sensor, error)
	Save(alert model.Sensor)
	SetAlert(sensor model.Sensor)
}

type SensorRepository struct {
	Db *gorm.DB
}

func NewSensorRepository(Db *gorm.DB) SensorRepositoryInterface {
	return &SensorRepository{Db: Db}
}

func (a *SensorRepository) Save(sensor model.Sensor) {
	result := a.Db.Create(&sensor)
	helper.ErrorPanic(result.Error)
}

func (s *SensorRepository) FindBySensorId(sensorId int) (model.Sensor, error) {
	var sensor model.Sensor
	result := s.Db.Where("id = ?", sensorId).First(&sensor)
	if err := result.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			helper.ErrorPanic(result.Error)
		}
	}

	return sensor, result.Error
}

func (s *SensorRepository) SetAlert(sensor model.Sensor) {
	status := model.Status{
		Name: model.ALERT,
	}
	sensor.Status = status
	s.Save(sensor)
}
