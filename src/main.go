package main

import (
	"alarm-api/config"
	"alarm-api/controller"
	"alarm-api/helper"
	"alarm-api/model"
	"alarm-api/repository"
	"alarm-api/router"
	"alarm-api/service"

	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Started Server!")

	db := config.DatabaseConnection()

	db.AutoMigrate(&model.Alert{}, &model.Measurement{}, &model.Sensor{}, &model.Status{})
	validate := validator.New()

	alertRepository := repository.NewAlertRepository(db)
	alertService := service.New(alertRepository, validate)
	alertController := controller.NewAlertController(alertService)

	measurementRepository := repository.NewMeasurmentRepository(db)
	sensorRepository := repository.NewSensorRepository(db)
	measurementService := service.NewMeasurmentService(measurementRepository, sensorRepository, validate)
	measurementController := controller.NewMeasurementController(measurementService)

	routes := router.New(alertController, measurementController)

	server := &http.Server{
		Addr:    ":8080",
		Handler: routes,
	}

	err := server.ListenAndServe()
	helper.ErrorPanic(err)
}
