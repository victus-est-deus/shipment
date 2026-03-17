package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type StatusEventRepository struct {
	db *sql.DB
}

func NewStatusEventRepository(db *sql.DB) *StatusEventRepository {
	return &StatusEventRepository{db: db}
}

func (r *StatusEventRepository) Create(ctx context.Context, e *entity.StatusEvent) error {
	query := `
		INSERT INTO status_events (id, shipment_id, status, location, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.ShipmentID, e.Status.String(), e.Location, e.Notes, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting status event: %w", err)
	}
	return nil
}

func (r *StatusEventRepository) GetByShipmentID(ctx context.Context, shipmentID uuid.UUID) ([]*entity.StatusEvent, error) {
	query := `
		SELECT id, shipment_id, status, location, notes, created_at
		FROM status_events WHERE shipment_id = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("querying status events: %w", err)
	}
	defer rows.Close()

	var events []*entity.StatusEvent
	for rows.Next() {
		var e entity.StatusEvent
		var statusStr string

		if err := rows.Scan(&e.ID, &e.ShipmentID, &statusStr, &e.Location, &e.Notes, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning status event: %w", err)
		}

		status, err := valueobject.ParseStatus(statusStr)
		if err != nil {
			return nil, err
		}
		e.Status = status

		events = append(events, &e)
	}

	return events, rows.Err()
}
