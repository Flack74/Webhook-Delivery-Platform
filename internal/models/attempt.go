package models

import (
	"encoding/json"
	"time"
)

type DeliveryAttempt struct {
	ID              string          `db:"id" json:"id"`
	DeliveryID      string          `db:"delivery_id" json:"delivery_id"`
	AttemptNumber   int             `db:"attempt_number" json:"attempt_number"`
	RequestHeaders  json.RawMessage `db:"request_headers" json:"request_headers"`
	RequestBody     json.RawMessage `db:"request_body" json:"request_body"`
	ResponseStatus  *int            `db:"response_status" json:"response_status"`
	ResponseHeaders json.RawMessage `db:"response_headers" json:"response_headers"`
	ResponseBody    *string         `db:"response_body" json:"response_body"`
	ErrorMessage    *string         `db:"error_message" json:"error_message"`
	DurationMs      int             `db:"duration_ms" json:"duration_ms"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
}
