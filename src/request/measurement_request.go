package request

type CreateMeasurementRequest struct {
	Level    uint `validate:"required" json:"level"`
	SensorId uint `validate:"required" json:"sensor_id"`
}
