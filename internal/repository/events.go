package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
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

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Insert(
	ctx context.Context,
	eventType string,
	payload []byte,
) (string, error) {

	id := uuid.NewString()

	_, err := r.db.Exec(ctx, `
			INSERT INTO events (id, event_type, payload)
			VALUES ($1, $2, $3)
		`, id, eventType, payload)

	if err != nil {
		return "", err
	}

	return id, nil
}
