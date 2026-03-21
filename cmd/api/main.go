package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/handler"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/postgres"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/redis"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not loaded, using system env")
	}
	if os.Getenv("GIN_MODE") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// SetUp Posgtres connection
	pool, err := postgres.NewPool()
	if err != nil {
		log.Fatalf("failed to connect to postgres %v", err)
	}
	defer pool.Close() // pool is closed when the application finishes

	fmt.Println("Successfully connected to the database!")

	// Setup redis connection
	redisClient, err := redis.NewClient()
	if err != nil {
		log.Fatalf("failed to connect to redis %v", err)
	}

	// Initialize Event repository
	applicationRepo := repository.NewApplicationRepository(pool)
	apiKeyRepo := repository.NewAPIKeyRepository(pool)
	endpointRepo := repository.NewEndpointRepository(pool)
	eventRepo := repository.NewEventRepository(pool)
	deliveryRepo := repository.NewDeliveryRepository(pool)

	// Create queue abstraction
	deliveryQueue := queue.NewRedisQueue(redisClient, "deliveries_queue", "processing_queue", "deliveries_delayed")

	if err := deliveryQueue.RecoverStalled(ctx); err != nil {
		log.Printf("recover stalled deliveries failed: %v", err)
	}

	//  Initialize Event handler
	applicationHandler := handler.NewApplicationHandler(applicationRepo)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyRepo)
	endpointHandler := handler.NewEndpointHandler(endpointRepo)
	eventHandler := handler.NewEventHandler(endpointRepo, eventRepo, deliveryRepo, deliveryQueue)

	// Initialize Worker Pool (3 workers)
	for range 3 {
		deliveryWorker := worker.NewWorker(deliveryRepo, eventRepo, endpointRepo, deliveryQueue)
		go deliveryWorker.Start(ctx)
	}

	// Set up the routes
	handler.SetUpRoutes(router, applicationHandler, apiKeyHandler, endpointHandler, eventHandler, apiKeyRepo)

	srv := &http.Server{
		Addr:              ":8000",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Initializing the server in goroutine so that
	// it won't block the graceful shutdown handling
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Listen to interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown
	stop()
	log.Println("shutting down gracefully, press ctrl+c again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// The request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown: ", err)
	}
	log.Println("Server exiting")
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
