package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/models"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/queue"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/repository"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/utils"
)

type Worker struct {
	deliveryRepo *repository.DeliveryRepository
	eventRepo    *repository.EventRepository
	endpointRepo *repository.EndpointRepository
	queue        *queue.RedisQueue
	httpClient   *http.Client
}

type deliveryResult struct {
	requestHeaders  json.RawMessage
	requestBody     json.RawMessage
	responseStatus  *int
	responseHeaders json.RawMessage
	responseBody    string
	errMessage      string
	duration        time.Duration
}

func NewWorker(
	deliveryRepo *repository.DeliveryRepository,
	eventRepo *repository.EventRepository,
	endpointRepo *repository.EndpointRepository,
	queue *queue.RedisQueue,
) *Worker {
	return &Worker{
		deliveryRepo: deliveryRepo,
		eventRepo:    eventRepo,
		endpointRepo: endpointRepo,
		queue:        queue,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return
		default:
		}

		if err := w.queue.PromoteDue(ctx, time.Now().UTC(), 100); err != nil {
			log.Printf("promote due retries failed: %v", err)
		}

		deliveryID, err := w.queue.Reserve(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("reserve error: %v", err)
			continue
		}

		if err = w.deliveryRepo.MarkInProgress(ctx, deliveryID); err != nil {
			log.Printf("mark in_progress failed for %s: %v", deliveryID, err)
			_ = w.queue.Nack(ctx, deliveryID)
			continue
		}

		delivery, err := w.deliveryRepo.GetByID(ctx, deliveryID)
		if err != nil {
			log.Printf("load delivery failed for %s: %v", deliveryID, err)
			_ = w.queue.Ack(ctx, deliveryID)
			continue
		}

		event, err := w.eventRepo.GetByID(ctx, delivery.EventID)
		if err != nil {
			w.failTerminal(ctx, deliveryID, fmt.Sprintf("load event: %v", err))
			continue
		}

		endpoint, err := w.endpointRepo.GetByID(ctx, delivery.EndpointID)
		if err != nil {
			w.failTerminal(ctx, deliveryID, fmt.Sprintf("load endpoint: %v", err))
			continue
		}

		result := w.sendHTTP(ctx, delivery.ID, event.ID, event.EventType, event.Payload, endpoint.URL, endpoint.Secret)
		attemptNumber := delivery.AttemptCount + 1

		if err := w.deliveryRepo.InsertAttempt(ctx, repository.DeliveryAttemptInput{
			DeliveryID:      delivery.ID,
			AttemptNumber:   attemptNumber,
			RequestHeaders:  result.requestHeaders,
			RequestBody:     result.requestBody,
			ResponseStatus:  result.responseStatus,
			ResponseHeaders: result.responseHeaders,
			ResponseBody:    result.responseBody,
			ErrorMessage:    result.errMessage,
			DurationMs:      int(result.duration.Milliseconds()),
		}); err != nil {
			log.Printf("insert attempt failed for %s: %v", deliveryID, err)
		}

		if result.errMessage == "" {
			if err := w.deliveryRepo.MarkSucceeded(ctx, deliveryID); err != nil {
				log.Printf("mark succeeded failed for %s: %v", deliveryID, err)
				_ = w.queue.Nack(ctx, deliveryID)
				continue
			}
			if err := w.queue.Ack(ctx, deliveryID); err != nil {
				log.Printf("ack failed for %s: %v", deliveryID, err)
			}
			continue
		}

		retryAt, retryable := nextRetryAt(*delivery, result.responseStatus, result.errMessage)
		if retryable {
			if err := w.deliveryRepo.ScheduleRetry(ctx, deliveryID, retryAt, result.errMessage); err != nil {
				log.Printf("schedule retry failed for %s: %v", deliveryID, err)
				_ = w.queue.Nack(ctx, deliveryID)
				continue
			}
			if err := w.queue.NackDelayed(ctx, deliveryID, retryAt); err != nil {
				log.Printf("enqueue delayed retry failed for %s: %v", deliveryID, err)
			}
			continue
		}

		if delivery.AttemptCount+1 >= delivery.MaxAttempts {
			if err := w.deliveryRepo.MarkDeadLetter(ctx, deliveryID, result.errMessage); err != nil {
				log.Printf("mark dead letter failed for %s: %v", deliveryID, err)
				_ = w.queue.Nack(ctx, deliveryID)
				continue
			}
		} else {
			if err := w.deliveryRepo.MarkFailed(ctx, deliveryID, result.errMessage); err != nil {
				log.Printf("mark failed failed for %s: %v", deliveryID, err)
				_ = w.queue.Nack(ctx, deliveryID)
				continue
			}
		}

		if err := w.queue.Ack(ctx, deliveryID); err != nil {
			log.Printf("ack after terminal failure failed for %s: %v", deliveryID, err)
		}
	}
}

func (w *Worker) failTerminal(ctx context.Context, deliveryID string, reason string) {
	if err := w.deliveryRepo.MarkFailed(ctx, deliveryID, reason); err != nil {
		log.Printf("mark terminal failure failed for %s: %v", deliveryID, err)
		_ = w.queue.Nack(ctx, deliveryID)
		return
	}
	if err := w.queue.Ack(ctx, deliveryID); err != nil {
		log.Printf("ack after terminal failure failed for %s: %v", deliveryID, err)
	}
}

func (w *Worker) sendHTTP(
	ctx context.Context,
	deliveryID string,
	eventID string,
	eventType string,
	payload []byte,
	endpointURL string,
	endpointSecret string,
) deliveryResult {
	startedAt := time.Now()
	timestamp := strconv.FormatInt(startedAt.UTC().Unix(), 10)
	signaturePayload := timestamp + "." + string(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(payload))
	if err != nil {
		return deliveryResult{
			requestBody: payload,
			errMessage:  fmt.Sprintf("build request: %v", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "webhook-delivery-platform/1.0")
	req.Header.Set("X-Webhook-Delivery-ID", deliveryID)
	req.Header.Set("X-Webhook-Event-ID", eventID)
	req.Header.Set("X-Webhook-Event-Type", eventType)
	req.Header.Set("X-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Webhook-Signature", utils.GenerateHMAC(signaturePayload, endpointSecret))

	requestHeaders, _ := json.Marshal(req.Header)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return deliveryResult{
			requestHeaders: requestHeaders,
			requestBody:    payload,
			errMessage:     classifyNetworkError(err),
			duration:       time.Since(startedAt),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	responseHeaders, _ := json.Marshal(resp.Header)
	responseBody := string(body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return deliveryResult{
			requestHeaders:  requestHeaders,
			requestBody:     payload,
			responseStatus:  &resp.StatusCode,
			responseHeaders: responseHeaders,
			responseBody:    responseBody,
			errMessage:      fmt.Sprintf("upstream responded with status %d", resp.StatusCode),
			duration:        time.Since(startedAt),
		}
	}

	return deliveryResult{
		requestHeaders:  requestHeaders,
		requestBody:     payload,
		responseStatus:  &resp.StatusCode,
		responseHeaders: responseHeaders,
		responseBody:    responseBody,
		duration:        time.Since(startedAt),
	}
}

func nextRetryAt(delivery models.Delivery, status *int, errMessage string) (time.Time, bool) {
	if delivery.AttemptCount+1 >= delivery.MaxAttempts {
		return time.Time{}, false
	}

	if status != nil {
		switch {
		case *status == http.StatusTooManyRequests,
			*status == http.StatusRequestTimeout,
			*status == http.StatusConflict,
			*status == http.StatusTooEarly:
		case *status >= 500:
		default:
			return time.Time{}, false
		}
	} else if errMessage == "" {
		return time.Time{}, false
	}

	return time.Now().UTC().Add(retryDelay(delivery.AttemptCount + 1)), true
}

func retryDelay(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	base := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
	if base > 15*time.Minute {
		base = 15 * time.Minute
	}

	jitter := time.Duration(float64(base) * 0.2)
	return base + jitter
}

func classifyNetworkError(err error) string {
	if err == nil {
		return ""
	}

	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) && timeout.Timeout() {
		return fmt.Sprintf("request timeout: %v", err)
	}

	return strings.TrimSpace(err.Error())
}
