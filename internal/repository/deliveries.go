package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Delivery struct {
	ID          string
	EventID     string
	EndpointURL string
	Status      string
	Attempts    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeliveryRepository struct {
	db *pgxpool.Pool
}

func NewDeliveryRepository(db *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{db: db}
}

var ErrDeliveryNotFound = errors.New("delivery not found")

func (d *DeliveryRepository) Insert(
	ctx context.Context,
	eventID string,
	endpointURL string,
) (string, error) {
	var id string

	err := d.db.QueryRow(ctx, `
		INSERT INTO deliveries (event_id, endpoint_url, status)
		VALUES ($1, $2, $3)
		RETURNING id
	`, eventID, endpointURL, "pending").Scan(&id)

	if err != nil {
		return "", err
	}
	return id, nil
}

func (d *DeliveryRepository) MarkInProgress(ctx context.Context, deliveryID string) error {
	cmdTag, err := d.db.Exec(ctx, `
			UPDATE deliveries
			SET status = 'in_progress',
				updated_at = now()
			WHERE id = $1
		`, deliveryID)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}
	return nil
}

func (d *DeliveryRepository) GetByID(ctx context.Context, deliveryID string) (*Delivery, error) {
	var newDelivery Delivery

	err := d.db.QueryRow(ctx, `
		SELECT id, event_id, endpoint_url, status, attempts, created_at, updated_at
		FROM deliveries
		WHERE id = $1
	`, deliveryID).Scan(
		&newDelivery.ID,
		&newDelivery.EventID,
		&newDelivery.EndpointURL,
		&newDelivery.Status,
		&newDelivery.Attempts,
		&newDelivery.CreatedAt,
		&newDelivery.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeliveryNotFound
		}
		return nil, err
	}

	return &newDelivery, nil
}

func (d *DeliveryRepository) MarkFailed(ctx context.Context, deliveryID string) error {
	cmdTag, err := d.db.Exec(ctx, `
			UPDATE deliveries
			SET status = 'failed',
				attempts = attempts + 1,
				updated_at = now()
			WHERE id = $1
		`, deliveryID)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}
	return nil
}

func (d *DeliveryRepository) MarkSucceeded(ctx context.Context, deliveryID string) error {
	cmdTag, err := d.db.Exec(ctx, `
		UPDATE deliveries
		SET status = 'succeeded',
		updated_at = now()
	where id = $1
	`, deliveryID)

	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrDeliveryNotFound
	}

	return nil
}
