package models

import "time"

type APIKey struct {
	ID            string    `db:"id" json:"id"`
	ApplicationID string    `db:"application_id" json:"application_id"`
	KeyHash       string    `db:"key_hash" json:"-"`
	Name          string    `db:"name" json:"name"`
	LastUsedAt    time.Time `db:"last_used_at" json:"last_used_at"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
