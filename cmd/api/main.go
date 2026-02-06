package main

import (
	"bytes"
	"encoding/json"
	"net/http"

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

var events []Event // Temporary in-memory storage (Phase 0 only)

func main() {
	router := gin.Default()

	// Register the POST endpoint
	router.POST("/events", createEvent)
	// Run the server
	router.Run("localhost:8000")
}

// Handler function for the POST /events endpoint
func createEvent(c *gin.Context) {
	var newEvent Event

	// Bind the incoming JSON to the newEvent struct.
	// Gin automatically validates based on "binding:required" tags.
	if err := c.ShouldBindJSON(&newEvent); err != nil {
		// If binding fails, return a 400 Bad Request error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Will move to database.
	events = append(events, newEvent)

	// Return a 202 Accepted status with new event data
	c.JSON(http.StatusAccepted, gin.H{
		"status": "accepted",
	})

	// Trigger Webhook
	deliverWebhook(newEvent)
}

func deliverWebhook(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	targetURL := "http://localhost:9000/webhook"
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	req.Header.Set("Control-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
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
