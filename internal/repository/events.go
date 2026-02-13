package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	ID        string
	EventType string
	Payload   []byte
	CreatedAt time.Time
}

type EventRepository struct {
	db *pgxpool.Pool
}

var ErrEventNotFound = errors.New("event not found")

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Insert(
	ctx context.Context,
	eventType string,
	payload []byte,
) (string, error) {

	var id string
	err := r.db.QueryRow(ctx, `
			INSERT INTO events (event_type, payload)
			VALUES ($1, $2)
			RETURNING id
		`, eventType, payload).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *EventRepository) GetByID(ctx context.Context, eventID string) (*Event, error) {
	var newEvent Event

	err := r.db.QueryRow(ctx, `
		SELECT id, event_type, payload, created_at
		FROM events
		WHERE id = $1
	`, eventID).Scan(
		&newEvent.ID,
		&newEvent.EventType,
		&newEvent.Payload,
		&newEvent.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}

	return &newEvent, nil
}
