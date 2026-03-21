package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type CreateEventRequest struct {
	EventType string          `json:"event_type" binding:"required,min=3,max=255"`
	Data      json.RawMessage `json:"data" binding:"required"`
}

type CreateEventResponse struct {
	Status           string `json:"status"`
	EventID          string `json:"event_id"`
	DeliveriesQueued int    `json:"deliveries_queued"`
}

type EventHandler struct {
	endpointRepo  *repository.EndpointRepository
	eventRepo     *repository.EventRepository
	deliveryRepo  *repository.DeliveryRepository
	deliveryQueue *queue.RedisQueue
}

func NewEventHandler(
	endpointRepo *repository.EndpointRepository,
	eventRepo *repository.EventRepository,
	deliveryRepo *repository.DeliveryRepository,
	deliveryQueue *queue.RedisQueue,
) *EventHandler {
	return &EventHandler{
		endpointRepo:  endpointRepo,
		eventRepo:     eventRepo,
		deliveryRepo:  deliveryRepo,
		deliveryQueue: deliveryQueue,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	applicationID, ok := authenticatedApplicationID(c)
	if !ok {
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		respondError(c, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	eventID, err := h.eventRepo.Insert(
		c.Request.Context(),
		applicationID,
		req.EventType,
		req.Data,
		idempotencyKey,
	)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	endpoints, err := h.endpointRepo.ListActiveByApplicationID(c.Request.Context(), applicationID)
	if err != nil {
		respondRepositoryError(c, err)
		return
	}

	enqueuedCount := 0
	for _, endpoint := range endpoints {
		deliveryID, err := h.deliveryRepo.Insert(c.Request.Context(), eventID, endpoint.ID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryAlreadyExists) {
				continue
			}
			respondRepositoryError(c, err)
			return
		}

		if err := h.deliveryQueue.Enqueue(c.Request.Context(), deliveryID); err != nil {
			respondError(c, http.StatusInternalServerError, "failed to enqueue delivery")
			return
		}

		enqueuedCount++
	}

	c.JSON(http.StatusAccepted, CreateEventResponse{
		Status:           "accepted",
		EventID:          eventID,
		DeliveriesQueued: enqueuedCount,
	})
}
