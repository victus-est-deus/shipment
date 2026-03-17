package jsonfile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type StatusEventRepository struct {
	store *Store
}

func NewStatusEventRepository(store *Store) *StatusEventRepository {
	return &StatusEventRepository{store: store}
}

func (r *StatusEventRepository) Create(_ context.Context, e *entity.StatusEvent) error {
	return r.store.SaveStatusEvent(StatusEventRecord{
		ID:         e.ID.String(),
		ShipmentID: e.ShipmentID.String(),
		Status:     e.Status.String(),
		Location:   e.Location,
		Notes:      e.Notes,
		CreatedAt:  TimeToString(e.CreatedAt),
	})
}

func (r *StatusEventRepository) GetByShipmentID(_ context.Context, shipmentID uuid.UUID) ([]*entity.StatusEvent, error) {
	records, err := r.store.GetStatusEventsByShipmentID(shipmentID)
	if err != nil {
		return nil, err
	}

	events := make([]*entity.StatusEvent, 0, len(records))
	for _, rec := range records {
		e, err := recordToStatusEvent(&rec)
		if err != nil {
			return nil, fmt.Errorf("converting status event record: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func recordToStatusEvent(r *StatusEventRecord) (*entity.StatusEvent, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	shipmentID, err := uuid.Parse(r.ShipmentID)
	if err != nil {
		return nil, err
	}

	status, err := valueobject.ParseStatus(r.Status)
	if err != nil {
		return nil, err
	}

	createdAt, _ := StringToTime(r.CreatedAt)

	return &entity.StatusEvent{
		ID:         id,
		ShipmentID: shipmentID,
		Status:     status,
		Location:   r.Location,
		Notes:      r.Notes,
		CreatedAt:  createdAt,
	}, nil
}
