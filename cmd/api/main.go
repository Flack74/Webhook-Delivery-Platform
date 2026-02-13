package main

import (
	"context"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/handler"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/redis"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	router := gin.Default()
	godotenv.Load()

	// SetUp Posgtres connection
	pool := repository.SetUpPostgresConn()
	defer pool.Close() // pool is closed when the application finishes

	// Setup redis connection
	redisClient := redis.RedisClient()

	// Initialize Event repository
	eventRepo := repository.NewEventRepository(pool)
	deliveryRepo := repository.NewDeliveryRepository(pool)

	// Create queue abstraction
	deliveryQueue := queue.NewRedisQueue(redisClient, "deliveries_queue", "processing_queue")

	//  Initialize Event handler
	eventHandler := handler.NewEventHandler(eventRepo, deliveryRepo, deliveryQueue)

	// Initialize Worker
	worker := worker.NewWorker(deliveryRepo, eventRepo, deliveryQueue)
	go worker.Start(context.Background())

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
