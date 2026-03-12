package models

import "time"

type Endpoint struct {
	ID            string    `db:"id" json:"id"`
	ApplicationID string    `db:"application_id" json:"application_id"`
	URL           string    `db:"url" json:"url"`
	Secret        string    `db:"secret" json:"-"`
	Description   string    `db:"description" json:"description"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	RateLimit     int       `db:"rate_limit" json:"rate_limit"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
