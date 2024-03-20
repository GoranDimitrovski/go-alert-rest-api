package repository

import (
	"alarm-api/helper"
	"alarm-api/model"

	"gorm.io/gorm"
)

type MeasurmentRepositoryInterface interface {
	FindAll(sensorId int) []model.Measurement
	Save(alert model.Measurement)
	IsAlert(sensorId int) bool
}

type MeasurmentRepository struct {
	Db *gorm.DB
}

func NewMeasurmentRepository(Db *gorm.DB) MeasurmentRepositoryInterface {
	return &MeasurmentRepository{Db: Db}
}

func (m *MeasurmentRepository) Save(measurment model.Measurement) {
	result := m.Db.Create(&measurment)
	helper.ErrorPanic(result.Error)
}

func (m *MeasurmentRepository) FindAll(sensorId int) []model.Measurement {
	var measurments []model.Measurement
	result := m.Db.Where("sensor_id = ?", sensorId).Find(&measurments)
	helper.ErrorPanic(result.Error)
	return measurments
}

func (m *MeasurmentRepository) IsAlert(sensorId int) bool {
	var count int64
	result := m.Db.Where("sensor_id = ?", sensorId).Count(&count)
	
	if result.Error != nil {
		panic(result.Error)
	}

	helper.ErrorPanic(result.Error)

	if count >= 3 {
		return true
	} else {
		return false
	}
}