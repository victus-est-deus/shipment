package jsonfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type ShipmentRepository struct {
	store *Store
}

func NewShipmentRepository(store *Store) *ShipmentRepository {
	return &ShipmentRepository{store: store}
}

func (r *ShipmentRepository) Create(_ context.Context, s *entity.Shipment) error {
	return r.store.SaveShipment(ShipmentRecord{
		ID:              s.ID.String(),
		ReferenceNumber: s.ReferenceNumber,
		Origin:          s.Origin,
		Destination:     s.Destination,
		Status:          s.Status.String(),
		DriverName:      s.DriverName,
		DriverPhone:     s.DriverPhone,
		UnitNumber:      s.UnitNumber,
		ShipmentAmount:  s.ShipmentAmount.Cents(),
		DriverRevenue:   s.DriverRevenue.Cents(),
		CreatedAt:       TimeToString(s.CreatedAt),
		UpdatedAt:       TimeToString(s.UpdatedAt),
	})
}

func (r *ShipmentRepository) GetByID(_ context.Context, id uuid.UUID) (*entity.Shipment, error) {
	record, err := r.store.GetShipment(id)
	if err != nil {
		return nil, fmt.Errorf("shipment not found: %w", err)
	}
	return recordToShipment(record)
}

func (r *ShipmentRepository) GetByReferenceNumber(_ context.Context, refNumber string) (*entity.Shipment, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	dir := filepath.Join(r.store.basePath, "shipments")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var record ShipmentRecord
		if err := r.store.readJSON(filepath.Join(dir, entry.Name()), &record); err != nil {
			continue
		}
		if record.ReferenceNumber == refNumber {
			return recordToShipment(&record)
		}
	}

	return nil, errors.New("shipment not found")
}

func (r *ShipmentRepository) Update(_ context.Context, s *entity.Shipment) error {
	return r.store.SaveShipment(ShipmentRecord{
		ID:              s.ID.String(),
		ReferenceNumber: s.ReferenceNumber,
		Origin:          s.Origin,
		Destination:     s.Destination,
		Status:          s.Status.String(),
		DriverName:      s.DriverName,
		DriverPhone:     s.DriverPhone,
		UnitNumber:      s.UnitNumber,
		ShipmentAmount:  s.ShipmentAmount.Cents(),
		DriverRevenue:   s.DriverRevenue.Cents(),
		CreatedAt:       TimeToString(s.CreatedAt),
		UpdatedAt:       TimeToString(s.UpdatedAt),
	})
}

func recordToShipment(r *ShipmentRecord) (*entity.Shipment, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	status, err := valueobject.ParseStatus(r.Status)
	if err != nil {
		return nil, err
	}

	amount, _ := valueobject.NewMoneyFromCents(r.ShipmentAmount, "USD")
	revenue, _ := valueobject.NewMoneyFromCents(r.DriverRevenue, "USD")

	createdAt, _ := StringToTime(r.CreatedAt)
	updatedAt, _ := StringToTime(r.UpdatedAt)

	return &entity.Shipment{
		ID:              id,
		ReferenceNumber: r.ReferenceNumber,
		Origin:          r.Origin,
		Destination:     r.Destination,
		Status:          status,
		DriverName:      r.DriverName,
		DriverPhone:     r.DriverPhone,
		UnitNumber:      r.UnitNumber,
		ShipmentAmount:  amount,
		DriverRevenue:   revenue,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}
