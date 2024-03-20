package controller

import (
	"alarm-api/service"
	"fmt"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
)

type SensorController struct {
	sensorService service.SensorServiceInterface
}

func NewSensorController(service service.SensorServiceInterface) *SensorController {
	return &SensorController{
		sensorService: service,
	}
}

func (controller *SensorController) GetSensor(ctx *gin.Context) {
    sensorId := ctx.Param("sensor_id")
    sensorIdInt, err := strconv.Atoi(sensorId)
    if err != nil {
        fmt.Println("Error:", err)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
        return
    }

    sensorResponse, err := controller.sensorService.FindBySensorId(sensorIdInt)
    if err != nil {
        ctx.JSON(http.StatusNotFound, nil)
        return
    }

    ctx.JSON(http.StatusOK, sensorResponse)
}
