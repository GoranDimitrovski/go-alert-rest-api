package router

import (
	"alarm-api/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(alertController *controller.AlertController, measurementController * controller.MeasurementController) *gin.Engine {
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	group := router.Group("/v1")
	{
		group.GET("/:sensor_id", alertController.GetAlerts)
		group.GET("/:sensor_id/alert", alertController.GetAlerts)
		group.GET("/:sensor_id/measurement", measurementController.GetMeasurements)
		group.POST("measurement", measurementController.Create)
	}
	
	return router
}
