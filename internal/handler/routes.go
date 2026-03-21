package handler

import (
	"github.com/Flack74/Webhook-Delivery-Platform/internal/middleware"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes(
	router *gin.Engine,
	applicationHandler *ApplicationHandler,
	apiKeyHandler *APIKeyHandler,
	endpointHandler *EndpointHandler,
	eventHandler *EventHandler,
	apiKeyRepo *repository.APIKeyRepository,
) {
	v1 := router.Group("/v1")

	v1.POST("/applications", applicationHandler.CreateApplication)

	protected := v1.Group("/")
	protected.Use(middleware.AuthMiddleware(apiKeyRepo))
	{
		protected.POST("/applications/:id/api-keys", apiKeyHandler.CreateAPIKey)
		protected.POST("/applications/:id/endpoints", endpointHandler.CreateEndpoint)
		protected.POST("/events", eventHandler.CreateEvent)
	}
}
