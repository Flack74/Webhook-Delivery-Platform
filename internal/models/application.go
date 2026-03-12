package models

import "time"

type Application struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	APIKeyHash string    `db:"api_key_hash" json:"-"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	IsActive   bool      `db:"is_active" json:"is_active"`
}
