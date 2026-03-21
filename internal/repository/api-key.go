package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Flack74/Webhook-Delivery-Platform/internal/models"
	"github.com/Flack74/Webhook-Delivery-Platform/internal/utils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKeyRepository struct {
	db *pgxpool.Pool
}

var ErrAPIKeyNotFound = errors.New("api-key not found")

func NewAPIKeyRepository(db *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Insert(
	ctx context.Context,
	applicationID string,
	rawKey string,
	name string,
) (string, error) {
	hashed := utils.HashAPIKey(rawKey)

	query := `
		INSERT INTO api_keys (application_id, key_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var apiKeyID string
	if err := r.db.QueryRow(ctx, query, applicationID, hashed, name).Scan(&apiKeyID); err != nil {
		return "", fmt.Errorf("insert api key: %w", err)
	}

	return apiKeyID, nil
}

func (r *APIKeyRepository) GetByKey(ctx context.Context, rawKey string) (*models.APIKey, error) {
	hashed := utils.HashAPIKey(rawKey)

	query := `
		SELECT id, application_id, key_hash, name, last_used_at, created_at
		FROM api_keys
		WHERE key_hash = $1
	`

	var apiKey models.APIKey
	if err := pgxscan.Get(ctx, r.db, &apiKey, query, hashed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, fmt.Errorf("get api key by key_hash: %w", err)
	}

	return &apiKey, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE api_keys
		SET last_used_at = now()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("update api key last used: %w", err)
	}

	return nil
}
