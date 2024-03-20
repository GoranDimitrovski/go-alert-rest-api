package response

import "time"

type AlertResponse struct {
	Id           uint      `json:"id"`
	SensorId     uint      `json:"sensor_id"`
	Measurement1 uint      `json:"measurement1"`
	Measurement2 uint      `json:"measurement2"`
	Measurement3 uint      `json:"measurement3"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}