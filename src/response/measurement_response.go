package response

type MeasurementResponse struct {
	Id       uint `json:"id"`
	Level    uint `json:"level"`
	SensorId uint `json:"sensor_id"`
}
