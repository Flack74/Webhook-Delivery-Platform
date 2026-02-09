package main

import (
	"github.com/Flack74/Webhook-Delivery-Platform/internal/handler"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// SetUp Posgtres connection
	pool := repository.SetUpPostgresConn()
	defer pool.Close() // pool is closed when the application finishes

	//
	eventRepo := repository.NewEventRepository(pool)

	eventHandler := handler.NewEventHandler(eventRepo)
	// Set up the routes
	handler.SetUpRoutes(router, eventHandler)

	// Run the server
	router.Run("localhost:8000")
}

/*
			Payload
{
  "event_type": "user.created",
  "data": {
    "id": "usr_123",
    "email": "user@example.com"
  }
}
*/
