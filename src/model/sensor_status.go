package model

const (
	OK    = "OK"
	WARN  = "WARN"
	ALERT = "ALERT"
)

type Status struct {
	Id   uint
	Name string
}

func (a *Status) TableName() string {
	return "sensor_status"
}
