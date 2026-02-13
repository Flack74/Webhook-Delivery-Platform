package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type Event struct {
	EventType string `json:"event_type" binding:"required"`
	Data      Data   `json:"data" binding:"required"`
}

type Data struct {
	ID    string `json:"id" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type EventHandler struct {
	eventRepo     *repository.EventRepository
	deliveryRepo  *repository.DeliveryRepository
	deliveryQueue *queue.RedisQueue
}

func NewEventHandler(eventRepo *repository.EventRepository, deliveryRepo *repository.DeliveryRepository, deliveryQueue *queue.RedisQueue) *EventHandler {
	return &EventHandler{
		eventRepo:     eventRepo,
		deliveryRepo:  deliveryRepo,
		deliveryQueue: deliveryQueue,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var newEvent Event

	// Bind the incoming JSON to the newEvent struct.
	// Gin automatically validates based on "binding:required" tags.
	if err := c.ShouldBindJSON(&newEvent); err != nil {
		// If binding fails, return a 400 Bad Request error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Serialize the data to JSON format
	payload, err := json.Marshal(newEvent.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Will move to database.
	event_id, err := h.eventRepo.Insert(c, newEvent.EventType, payload)
	if err != nil {
		log.Printf("Unable to insert the payload to database!\n%v", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	endpointURL := "http://localhost:9000/webhook"

	// Push the event to the delivery queue
	deliveryID, err := h.deliveryRepo.Insert(c, event_id, endpointURL)

	if err != nil {
		log.Printf("Unable to insert the delivery to database!\n%v", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.deliveryQueue.Enqueue(c, deliveryID)

	if err != nil {
		log.Printf("Unable to enqueue the delivery to redis!\n%v", err)

		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	// Return a 202 Accepted status with new event data
	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"event_id": event_id,
	})
}
