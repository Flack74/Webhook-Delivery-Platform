package models

import "time"

type Delivery struct {
	ID           string     `db:"id" json:"id"`
	EventID      string     `db:"event_id" json:"event_id"`
	EndpointID   string     `db:"endpoint_id" json:"endpoint_id"`
	Status       string     `db:"status" json:"status"`
	AttemptCount int        `db:"attempt_count" json:"attempt_count"`
	NextRetryAt  *time.Time `db:"next_retry_at" json:"next_retry_at"` // can be null so * , otherwise scanning from DB may fail.
	LastError    *string    `db:"last_error" json:"last_error"`
	MaxAttempts  int        `db:"max_attempts" json:"max_attempts"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
