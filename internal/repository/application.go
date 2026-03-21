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

type ApplicationRepository struct {
	db *pgxpool.Pool
}

var ErrApplicationNotFound = errors.New("application not found")

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Insert(ctx context.Context, name string) (*models.Application, error) {
	var app models.Application

	query := `
		INSERT INTO applications (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at, is_active
	`

	if err := pgxscan.Get(ctx, r.db, &app, query, name); err != nil {
		return nil, fmt.Errorf("insert application: %w", err)
	}

	return &app, nil
}

func (r *ApplicationRepository) CreateWithAPIKey(
	ctx context.Context,
	name string,
	apiKeyHash string,
	apiKeyName string,
) (*models.Application, string, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("begin application transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var app models.Application
	if err := pgxscan.Get(ctx, tx, &app, `
		INSERT INTO applications (name)
		VALUES ($1)
		RETURNING id, name, created_at, updated_at, is_active
	`, name); err != nil {
		return nil, "", fmt.Errorf("insert application: %w", err)
	}

	var apiKeyID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO api_keys (application_id, key_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id
	`, app.ID, apiKeyHash, apiKeyName).Scan(&apiKeyID); err != nil {
		return nil, "", fmt.Errorf("insert bootstrap api key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", fmt.Errorf("commit application transaction: %w", err)
	}

	return &app, apiKeyID, nil
}

func (r *ApplicationRepository) GetByID(ctx context.Context, applicationID string) (*models.Application, error) {
	var app models.Application

	query := `
		SELECT id, name, created_at, updated_at, is_active
		FROM applications
		WHERE id = $1
	`

	if err := pgxscan.Get(ctx, r.db, &app, query, applicationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("get application by id: %w", err)
	}

	return &app, nil
}
