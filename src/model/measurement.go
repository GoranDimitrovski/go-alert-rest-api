package model

import (
	"time"
)

type Measurement struct {
	Id        uint
	Level     uint
	SensorID  uint
	Sensor    Sensor
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (a *Measurement) TableName() string {
	return "measurement"
}
