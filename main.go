package main

import (
	"fmt"

	controller "service_line_furk/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.Static("/static", "./image")
	router.GET("/debug", controller.ControllerDebug)
	router.POST("/api/line", controller.ControllerLineReplyMsg)

	err := router.Run(":1111")
	if err != nil {
		fmt.Println("error => ", err)
	}
}
