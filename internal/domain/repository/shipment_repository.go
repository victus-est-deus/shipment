package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
)

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *entity.Shipment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Shipment, error)
	GetByReferenceNumber(ctx context.Context, referenceNumber string) (*entity.Shipment, error)
	Update(ctx context.Context, shipment *entity.Shipment) error
}

type StatusEventRepository interface {
	Create(ctx context.Context, event *entity.StatusEvent) error
	GetByShipmentID(ctx context.Context, shipmentID uuid.UUID) ([]*entity.StatusEvent, error)
}

type LogRepository interface {
	Create(ctx context.Context, log *entity.Log) error
	GetByAction(ctx context.Context, action string) ([]*entity.Log, error)
}
