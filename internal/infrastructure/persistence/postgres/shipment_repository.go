package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type ShipmentRepository struct {
	db *sql.DB
}

func NewShipmentRepository(db *sql.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

func (r *ShipmentRepository) Create(ctx context.Context, s *entity.Shipment) error {
	query := `
		INSERT INTO shipments (id, reference_number, origin, destination, status, driver_name, driver_phone, unit_number, shipment_amount, driver_revenue, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.ReferenceNumber, s.Origin, s.Destination, s.Status.String(),
		s.DriverName, s.DriverPhone, s.UnitNumber,
		s.ShipmentAmount.Cents(), s.DriverRevenue.Cents(),
		s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting shipment: %w", err)
	}
	return nil
}

func (r *ShipmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Shipment, error) {
	query := `
		SELECT id, reference_number, origin, destination, status, driver_name, driver_phone, unit_number, shipment_amount, driver_revenue, created_at, updated_at
		FROM shipments WHERE id = $1`

	return r.scanShipment(r.db.QueryRowContext(ctx, query, id))
}

func (r *ShipmentRepository) GetByReferenceNumber(ctx context.Context, refNumber string) (*entity.Shipment, error) {
	query := `
		SELECT id, reference_number, origin, destination, status, driver_name, driver_phone, unit_number, shipment_amount, driver_revenue, created_at, updated_at
		FROM shipments WHERE reference_number = $1`

	return r.scanShipment(r.db.QueryRowContext(ctx, query, refNumber))
}

func (r *ShipmentRepository) Update(ctx context.Context, s *entity.Shipment) error {
	query := `
		UPDATE shipments SET status = $1, driver_name = $2, driver_phone = $3, unit_number = $4, shipment_amount = $5, driver_revenue = $6, updated_at = $7
		WHERE id = $8`

	_, err := r.db.ExecContext(ctx, query,
		s.Status.String(), s.DriverName, s.DriverPhone, s.UnitNumber,
		s.ShipmentAmount.Cents(), s.DriverRevenue.Cents(),
		s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("updating shipment: %w", err)
	}
	return nil
}

func (r *ShipmentRepository) scanShipment(row *sql.Row) (*entity.Shipment, error) {
	var s entity.Shipment
	var statusStr string
	var amountCents, revenueCents int64

	err := row.Scan(
		&s.ID, &s.ReferenceNumber, &s.Origin, &s.Destination, &statusStr,
		&s.DriverName, &s.DriverPhone, &s.UnitNumber,
		&amountCents, &revenueCents,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning shipment: %w", err)
	}

	status, err := valueobject.ParseStatus(statusStr)
	if err != nil {
		return nil, err
	}
	s.Status = status

	s.ShipmentAmount, _ = valueobject.NewMoneyFromCents(amountCents, "USD")
	s.DriverRevenue, _ = valueobject.NewMoneyFromCents(revenueCents, "USD")

	return &s, nil
}
