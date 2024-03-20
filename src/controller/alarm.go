package controller

import (
	"alarm-api/service"
	"fmt"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
)

type AlertController struct {
	alertService service.AlertServiceInterface
}

func NewAlertController(service service.AlertServiceInterface) *AlertController {
	return &AlertController{
		alertService: service,
	}
}

func (controller *AlertController) GetAlerts(ctx *gin.Context) {
	sensorId := ctx.Param("sensor_id")
	sensorIdInt, err := strconv.Atoi(sensorId)
	if err != nil {
		fmt.Println("Error:", err)
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	alertResponse := controller.alertService.FindAllBySensorId(sensorIdInt)
	ctx.JSON(http.StatusOK, alertResponse)
}
