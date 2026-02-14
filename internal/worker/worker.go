package worker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
)

// Worker processes pending

type Worker struct {
	deliveryRepo *repository.DeliveryRepository
	eventRepo    *repository.EventRepository
	queue        *queue.RedisQueue
}

func NewWorker(deliveryRepo *repository.DeliveryRepository, eventRepo *repository.EventRepository, queue *queue.RedisQueue) *Worker {
	return &Worker{
		deliveryRepo: deliveryRepo,
		eventRepo:    eventRepo,
		queue:        queue,
	}
}

func (w *Worker) Start(ctx context.Context) {

	for {

		deliveryID, err := w.queue.Reserve(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("reserve error: %v", err)
			continue
		}

		err = w.deliveryRepo.MarkInProgress(ctx, deliveryID)
		if err != nil {
			log.Printf("mark in_progress failed : %v", err)
			continue
		}

		delivery, err := w.deliveryRepo.GetByID(ctx, deliveryID)
		if err != nil {
			log.Printf("delivery not found: %v", err)
			continue
		}

		event, err := w.eventRepo.GetByID(ctx, delivery.EventID)
		if err != nil {
			_ = w.deliveryRepo.MarkFailed(ctx, deliveryID)
			continue
		}

		if err := w.sendHTTP(event.Payload, delivery.EndpointURL); err != nil {
			log.Printf("delivery failed: %v", err)
			_ = w.deliveryRepo.MarkFailed(ctx, deliveryID)
			_ = w.queue.Nack(ctx, deliveryID) // Return to queue for retry
			continue
		}

		if err := w.deliveryRepo.MarkSucceeded(ctx, deliveryID); err != nil {
			log.Printf("mark succeeded failed: %v", err)
			continue
		}

		if err := w.queue.Ack(ctx, deliveryID); err != nil {
			log.Printf("ack failed: %v", err)
		}
	}

}

func (w *Worker) sendHTTP(payload []byte, endpointURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", endpointURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
