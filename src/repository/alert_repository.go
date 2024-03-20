package repository

import (
	"alarm-api/model"
	"alarm-api/helper"

	"gorm.io/gorm"
)

type AlertRepositoryInterface interface {
	FindAllBySensorId(sensorId int) []model.Alert
	Save(alert model.Alert)
}

type AlertRepository struct {
	Db *gorm.DB
}

func NewAlertRepository(Db *gorm.DB) AlertRepositoryInterface {
	return &AlertRepository{Db: Db}
}

func (t *AlertRepository) Save(alert model.Alert) {
	result := t.Db.Create(&alert)
	helper.ErrorPanic(result.Error)
}

func (a *AlertRepository) FindAllBySensorId(sensorId int) []model.Alert {
	var alerts []model.Alert
	result := a.Db.Where("sensor_id = ?", sensorId).Find(&alerts)
	helper.ErrorPanic(result.Error)
	return alerts
}