package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeliveryRepository struct {
	db *pgxpool.Pool
}

type DeliveryAttemptInput struct {
	DeliveryID      string
	AttemptNumber   int
	RequestHeaders  json.RawMessage
	RequestBody     json.RawMessage
	ResponseStatus  *int
	ResponseHeaders json.RawMessage
	ResponseBody    string
	ErrorMessage    string
	DurationMs      int
}

func NewDeliveryRepository(db *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

var ErrDeliveryNotFound = errors.New("delivery not found")
var ErrDeliveryNotAvailable = errors.New("delivery not available for processing")
var ErrDeliveryAlreadyExists = errors.New("delivery already exists")

func (r *DeliveryRepository) Insert(
	ctx context.Context,
	eventID string,
	endpointID string,
) (string, error) {
	var id string

	err := r.db.QueryRow(ctx, `
		INSERT INTO deliveries (event_id, endpoint_id)
		VALUES ($1, $2)
		RETURNING id
	`, eventID, endpointID).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrDeliveryAlreadyExists
		}
		return "", fmt.Errorf("insert delivery: %w", err)
	}

	return id, nil
}

func (r *DeliveryRepository) MarkInProgress(ctx context.Context, deliveryID string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'in_progress',
			updated_at = now()
		WHERE id = $1
			AND status = 'pending'
			AND (next_retry_at IS NULL OR next_retry_at <= now())
	`, deliveryID)
	if err != nil {
		return fmt.Errorf("mark in progress: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotAvailable
	}

	return nil
}

func (r *DeliveryRepository) GetByID(ctx context.Context, deliveryID string) (*models.Delivery, error) {
	var delivery models.Delivery

	query := `
		SELECT id, event_id, endpoint_id, status, attempt_count, next_retry_at, last_error, max_attempts, created_at, updated_at
		FROM deliveries
		WHERE id = $1
	`

	if err := pgxscan.Get(ctx, r.db, &delivery, query, deliveryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeliveryNotFound
		}
		return nil, fmt.Errorf("get delivery by id: %w", err)
	}

	return &delivery, nil
}

func (r *DeliveryRepository) InsertAttempt(ctx context.Context, in DeliveryAttemptInput) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO delivery_attempts (
			delivery_id,
			attempt_number,
			request_headers,
			request_body,
			response_status,
			response_headers,
			response_body,
			error_message,
			duration_ms
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, in.DeliveryID, in.AttemptNumber, in.RequestHeaders, in.RequestBody, in.ResponseStatus, in.ResponseHeaders, in.ResponseBody, in.ErrorMessage, in.DurationMs)
	if err != nil {
		return fmt.Errorf("insert delivery attempt: %w", err)
	}

	return nil
}

func (r *DeliveryRepository) MarkSucceeded(ctx context.Context, deliveryID string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'succeeded',
			attempt_count = attempt_count + 1,
			next_retry_at = NULL,
			last_error = NULL,
			updated_at = now()
		WHERE id = $1 AND status = 'in_progress'
	`, deliveryID)
	if err != nil {
		return fmt.Errorf("mark succeeded: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}

	return nil
}

func (r *DeliveryRepository) ScheduleRetry(ctx context.Context, deliveryID string, nextRetryAt time.Time, lastError string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'pending',
			attempt_count = attempt_count + 1,
			next_retry_at = $2,
			last_error = $3,
			updated_at = now()
		WHERE id = $1 AND status = 'in_progress'
	`, deliveryID, nextRetryAt, lastError)
	if err != nil {
		return fmt.Errorf("schedule retry: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}

	return nil
}

func (r *DeliveryRepository) MarkFailed(ctx context.Context, deliveryID string, lastError string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'failed',
			attempt_count = attempt_count + 1,
			next_retry_at = NULL,
			last_error = $2,
			updated_at = now()
		WHERE id = $1 AND status = 'in_progress'
	`, deliveryID, lastError)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}

	return nil
}

func (r *DeliveryRepository) MarkDeadLetter(ctx context.Context, deliveryID string, lastError string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'dead_letter',
			attempt_count = attempt_count + 1,
			next_retry_at = NULL,
			last_error = $2,
			updated_at = now()
		WHERE id = $1 AND status = 'in_progress'
	`, deliveryID, lastError)
	if err != nil {
		return fmt.Errorf("mark dead letter: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}

	return nil
}
