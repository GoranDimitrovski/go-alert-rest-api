package model

type Sensor struct {
	Id       uint
	StatusId uint
	Status   Status
}

func (a *Sensor) TableName() string {
	return "sensor"
}
