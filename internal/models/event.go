package models

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID             string          `db:"id" json:"id"`
	ApplicationID  string          `db:"application_id" json:"application_id"`
	EventType      string          `db:"event_type" json:"event_type"`
	Payload        json.RawMessage `db:"payload" json:"payload"`
	IdempotencyKey string          `db:"idempotency_key" json:"idempotency_key"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}
