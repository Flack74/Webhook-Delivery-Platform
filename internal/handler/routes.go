package handler

import "github.com/gin-gonic/gin"

func SetUpRoutes(router *gin.Engine, handler *EventHandler) {
	router.POST("/events", handler.CreateEvent)
}
