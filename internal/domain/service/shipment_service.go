package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/repository"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

var (
	ErrShipmentNotFound     = errors.New("shipment not found")
	ErrDuplicateReference   = errors.New("shipment with this reference number already exists")
	ErrInvalidTransition    = errors.New("invalid status transition")
	ErrShipmentTerminated   = errors.New("shipment is in a terminal state")
)

type ShipmentService struct {
	shipmentRepo    repository.ShipmentRepository
	statusEventRepo repository.StatusEventRepository
	logRepo         repository.LogRepository
}

func NewShipmentService(
	shipmentRepo repository.ShipmentRepository,
	statusEventRepo repository.StatusEventRepository,
	logRepo repository.LogRepository,
) *ShipmentService {
	return &ShipmentService{
		shipmentRepo:    shipmentRepo,
		statusEventRepo: statusEventRepo,
		logRepo:         logRepo,
	}
}

func (s *ShipmentService) CreateShipment(
	ctx context.Context,
	referenceNumber string,
	origin string,
	destination string,
	driverName string,
	driverPhone string,
	unitNumber string,
	shipmentAmount valueobject.Money,
	driverRevenue valueobject.Money,
) (*entity.Shipment, error) {
	existing, _ := s.shipmentRepo.GetByReferenceNumber(ctx, referenceNumber)
	if existing != nil {
		return nil, ErrDuplicateReference
	}

	shipment, err := entity.NewShipment(
		referenceNumber,
		origin,
		destination,
		driverName,
		driverPhone,
		unitNumber,
		shipmentAmount,
		driverRevenue,
	)
	if err != nil {
		return nil, err
	}

	if err := s.shipmentRepo.Create(ctx, shipment); err != nil {
		return nil, err
	}

	initialEvent := entity.NewStatusEvent(shipment.ID, valueobject.StatusPending, origin, "shipment created")
	if err := s.statusEventRepo.Create(ctx, initialEvent); err != nil {
		return nil, err
	}

	s.logRepo.Create(ctx, entity.NewLog("shipment_created", map[string]any{
		"shipment_id":      shipment.ID.String(),
		"reference_number": referenceNumber,
	}))

	return shipment, nil
}

func (s *ShipmentService) GetShipment(ctx context.Context, id uuid.UUID) (*entity.Shipment, error) {
	shipment, err := s.shipmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrShipmentNotFound
	}
	return shipment, nil
}

func (s *ShipmentService) UpdateStatus(
	ctx context.Context,
	shipmentID uuid.UUID,
	newStatus valueobject.Status,
	location string,
	notes string,
) (*entity.StatusEvent, error) {
	shipment, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	if shipment.Status.IsTerminal() {
		return nil, ErrShipmentTerminated
	}

	event, err := shipment.AddStatusEvent(newStatus, location, notes)
	if err != nil {
		return nil, ErrInvalidTransition
	}

	if err := s.shipmentRepo.Update(ctx, shipment); err != nil {
		return nil, err
	}

	if err := s.statusEventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	s.logRepo.Create(ctx, entity.NewLog("status_updated", map[string]any{
		"shipment_id": shipmentID.String(),
		"from_status": shipment.Status.String(),
		"to_status":   newStatus.String(),
	}))

	return event, nil
}

func (s *ShipmentService) GetEventHistory(ctx context.Context, shipmentID uuid.UUID) ([]*entity.StatusEvent, error) {
	_, err := s.shipmentRepo.GetByID(ctx, shipmentID)
	if err != nil {
		return nil, ErrShipmentNotFound
	}

	return s.statusEventRepo.GetByShipmentID(ctx, shipmentID)
}
