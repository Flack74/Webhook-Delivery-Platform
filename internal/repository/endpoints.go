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

type EndpointRepository struct {
	db *pgxpool.Pool
}

var ErrEndpointNotFound = errors.New("endpoint not found")

func NewEndpointRepository(db *pgxpool.Pool) *EndpointRepository {
	return &EndpointRepository{db: db}
}

func (r *EndpointRepository) Insert(
	ctx context.Context,
	appID string,
	url string,
	secret string,
	description string,
) (string, error) {
	var id string

	err := r.db.QueryRow(ctx, `
		INSERT INTO endpoints(application_id, url, secret, description)
		VALUES($1, $2, $3, $4)
		RETURNING id
	`, appID, url, secret, description).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert endpoint: %w", err)
	}

	return id, nil
}

func (r *EndpointRepository) GetByID(ctx context.Context, endpointID string) (*models.Endpoint, error) {
	var endpoint models.Endpoint

	query := `
		SELECT id, application_id, url, secret, description, is_active, rate_limit, created_at, updated_at
		FROM endpoints
		WHERE id = $1
	`

	if err := pgxscan.Get(ctx, r.db, &endpoint, query, endpointID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEndpointNotFound
		}
		return nil, fmt.Errorf("get endpoint by id: %w", err)
	}

	return &endpoint, nil
}

func (r *EndpointRepository) ListActiveByApplicationID(ctx context.Context, appID string) ([]models.Endpoint, error) {
	var endpoints []models.Endpoint

	query := `
		SELECT id, application_id, url, secret, description, is_active, rate_limit, created_at, updated_at
		FROM endpoints
		WHERE application_id = $1 AND is_active = TRUE
		ORDER BY created_at ASC
	`

	if err := pgxscan.Select(ctx, r.db, &endpoints, query, appID); err != nil {
		return nil, fmt.Errorf("list endpoints by application_id: %w", err)
	}

	return endpoints, nil
}
