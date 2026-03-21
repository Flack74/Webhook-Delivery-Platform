package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/models"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

var ErrEventNotFound = errors.New("event not found")

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Insert(
	ctx context.Context,
	applicationID string,
	eventType string,
	payload []byte,
	idempotencyKey string,
) (string, error) {

	var id string
	err := r.db.QueryRow(ctx, `
			INSERT INTO events (event_type, payload, application_id, idempotency_key)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (application_id, idempotency_key) DO NOTHING
			RETURNING id
		`, eventType, payload, applicationID, idempotencyKey).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.getByIdempotency(ctx, applicationID, idempotencyKey)
		}
		return "", fmt.Errorf("insert event: %w", err)
	}

	return id, nil
}

func (r *EventRepository) GetByID(ctx context.Context, eventID string) (*models.Event, error) {
	var newEvent models.Event

	query := `
		SELECT id, application_id, event_type, payload, idempotency_key, created_at
	 	FROM events
	 	WHERE id = $1
	`

	err := pgxscan.Get(ctx, r.db, &newEvent, query, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("get event by id: %w", err)
	}

	return &newEvent, nil
}

func (r *EventRepository) getByIdempotency(
	ctx context.Context,
	applicationID string,
	idempotencyKey string,
) (string, error) {

	var id string

	err := r.db.QueryRow(ctx, `
		SELECT id FROM events
		WHERE application_id = $1 AND idempotency_key = $2
		`, applicationID, idempotencyKey).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("get event by idempotency: %w", err)
	}
	return id, nil
}
