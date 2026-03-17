package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type Shipment struct {
	ID              uuid.UUID
	ReferenceNumber string
	Origin          string
	Destination     string
	Status          valueobject.Status
	DriverName      string
	DriverPhone     string
	UnitNumber      string
	ShipmentAmount  valueobject.Money
	DriverRevenue   valueobject.Money
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewShipment(
	referenceNumber string,
	origin string,
	destination string,
	driverName string,
	driverPhone string,
	unitNumber string,
	shipmentAmount valueobject.Money,
	driverRevenue valueobject.Money,
) (*Shipment, error) {
	if referenceNumber == "" {
		return nil, errors.New("reference number must not be empty")
	}
	if origin == "" {
		return nil, errors.New("origin must not be empty")
	}
	if destination == "" {
		return nil, errors.New("destination must not be empty")
	}

	now := time.Now().UTC()

	return &Shipment{
		ID:              uuid.New(),
		ReferenceNumber: referenceNumber,
		Origin:          origin,
		Destination:     destination,
		Status:          valueobject.StatusPending,
		DriverName:      driverName,
		DriverPhone:     driverPhone,
		UnitNumber:      unitNumber,
		ShipmentAmount:  shipmentAmount,
		DriverRevenue:   driverRevenue,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (s *Shipment) AddStatusEvent(newStatus valueobject.Status, location string, notes string) (*StatusEvent, error) {
	if !newStatus.IsValid() {
		return nil, errors.New("invalid status")
	}
	if !s.Status.CanTransitionTo(newStatus) {
		return nil, errors.New("invalid status transition from " + s.Status.String() + " to " + newStatus.String())
	}

	event := NewStatusEvent(s.ID, newStatus, location, notes)

	s.Status = newStatus
	s.UpdatedAt = event.CreatedAt

	return event, nil
}
