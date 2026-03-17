package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type StatusEvent struct {
	ID         uuid.UUID
	ShipmentID uuid.UUID
	Status     valueobject.Status
	Location   string
	Notes      string
	CreatedAt  time.Time
}

func NewStatusEvent(shipmentID uuid.UUID, status valueobject.Status, location string, notes string) *StatusEvent {
	return &StatusEvent{
		ID:         uuid.New(),
		ShipmentID: shipmentID,
		Status:     status,
		Location:   location,
		Notes:      notes,
		CreatedAt:  time.Now().UTC(),
	}
}
