package model

import (
	"time"
)

type Alert struct {
	Id           uint
	SensorID     uint
	Sensor       Sensor
	Measurement1 uint
	Measurement2 uint
	Measurement3 uint
	StartTime    time.Time
	EndTime      time.Time
}

func (a *Alert) TableName() string {
	return "alert"
}
