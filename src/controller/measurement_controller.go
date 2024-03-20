package controller

import (
	"alarm-api/helper"
	"alarm-api/request"
	"alarm-api/service"
	"fmt"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type MeasurementController struct {
	measurementService service.MeasurementServiceInterface
}

func NewMeasurementController(service service.MeasurementServiceInterface) *MeasurementController {
	return &MeasurementController{
		measurementService: service,
	}
}

func (controller *MeasurementController) Create(ctx *gin.Context) {
	log.Info().Msg("create measurement")
	createTagsRequest := request.CreateMeasurementRequest{}
	err := ctx.ShouldBindJSON(&createTagsRequest)
	helper.ErrorPanic(err)

	controller.measurementService.Create(createTagsRequest)
	ctx.JSON(http.StatusOK, nil)
}

func (controller *MeasurementController) GetMeasurements(ctx *gin.Context) {
	sensorId := ctx.Param("sensor_id")
	log.Info().Msg("findAll measurements")
	sensorIdInt, err := strconv.Atoi(sensorId)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	measurementResponse := controller.measurementService.FindAll(sensorIdInt)
	ctx.JSON(http.StatusOK, measurementResponse)
}
