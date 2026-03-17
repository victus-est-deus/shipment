package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/application/dto"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/service"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

type ShipmentUseCase struct {
	service *service.ShipmentService
}

func NewShipmentUseCase(svc *service.ShipmentService) *ShipmentUseCase {
	return &ShipmentUseCase{service: svc}
}

func (uc *ShipmentUseCase) CreateShipment(ctx context.Context, req dto.CreateShipmentRequest) (*entity.Shipment, error) {
	shipmentAmount, err := req.ToShipmentAmount()
	if err != nil {
		return nil, err
	}

	driverRevenue, err := req.ToDriverRevenue()
	if err != nil {
		return nil, err
	}

	return uc.service.CreateShipment(
		ctx,
		req.ReferenceNumber,
		req.Origin,
		req.Destination,
		req.DriverName,
		req.DriverPhone,
		req.UnitNumber,
		shipmentAmount,
		driverRevenue,
	)
}

func (uc *ShipmentUseCase) GetShipment(ctx context.Context, id string) (*entity.Shipment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return uc.service.GetShipment(ctx, uid)
}

func (uc *ShipmentUseCase) UpdateStatus(ctx context.Context, req dto.UpdateStatusRequest) (*entity.StatusEvent, error) {
	shipmentID, err := uuid.Parse(req.ShipmentID)
	if err != nil {
		return nil, err
	}

	status, err := valueobject.ParseStatus(req.Status)
	if err != nil {
		return nil, err
	}

	return uc.service.UpdateStatus(ctx, shipmentID, status, req.Location, req.Notes)
}

func (uc *ShipmentUseCase) GetEventHistory(ctx context.Context, shipmentID string) ([]*entity.StatusEvent, error) {
	uid, err := uuid.Parse(shipmentID)
	if err != nil {
		return nil, err
	}
	return uc.service.GetEventHistory(ctx, uid)
}
