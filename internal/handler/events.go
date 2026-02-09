package handler

import (
	"encoding/json"
	"log"
	"net/http"

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
	eventRepo *repository.EventRepository
}

func NewEventHandler(eventRepo *repository.EventRepository) *EventHandler {
	return &EventHandler{
		eventRepo: eventRepo,
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

	// Return a 202 Accepted status with new event data
	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"event_id": event_id,
	})
}
