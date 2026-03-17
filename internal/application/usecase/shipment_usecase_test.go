package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victus-est-deus/shipment/internal/application/dto"
	"github.com/victus-est-deus/shipment/internal/application/usecase"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/service"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

// --- Mocks ---

type mockShipmentRepo struct {
	shipments map[uuid.UUID]*entity.Shipment
}

func newMockShipmentRepo() *mockShipmentRepo {
	return &mockShipmentRepo{shipments: make(map[uuid.UUID]*entity.Shipment)}
}

func (m *mockShipmentRepo) Create(ctx context.Context, s *entity.Shipment) error {
	m.shipments[s.ID] = s
	return nil
}

func (m *mockShipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Shipment, error) {
	if s, ok := m.shipments[id]; ok {
		return s, nil
	}
	return nil, service.ErrShipmentNotFound
}

func (m *mockShipmentRepo) GetByReferenceNumber(ctx context.Context, ref string) (*entity.Shipment, error) {
	for _, s := range m.shipments {
		if s.ReferenceNumber == ref {
			return s, nil
		}
	}
	return nil, service.ErrShipmentNotFound
}

func (m *mockShipmentRepo) Update(ctx context.Context, s *entity.Shipment) error {
	m.shipments[s.ID] = s
	return nil
}

type mockStatusEventRepo struct {
	events map[uuid.UUID][]*entity.StatusEvent
}

func newMockStatusEventRepo() *mockStatusEventRepo {
	return &mockStatusEventRepo{events: make(map[uuid.UUID][]*entity.StatusEvent)}
}

func (m *mockStatusEventRepo) Create(ctx context.Context, e *entity.StatusEvent) error {
	m.events[e.ShipmentID] = append(m.events[e.ShipmentID], e)
	return nil
}

func (m *mockStatusEventRepo) GetByShipmentID(ctx context.Context, id uuid.UUID) ([]*entity.StatusEvent, error) {
	return m.events[id], nil
}

type mockLogRepo struct{}

func (m *mockLogRepo) Create(ctx context.Context, l *entity.Log) error           { return nil }
func (m *mockLogRepo) GetByAction(ctx context.Context, action string) ([]*entity.Log, error) { return nil, nil }

// --- Tests ---

func TestShipmentUseCase_CreateShipment(t *testing.T) {
	svc := service.NewShipmentService(newMockShipmentRepo(), newMockStatusEventRepo(), &mockLogRepo{})
	uc := usecase.NewShipmentUseCase(svc)

	req := dto.CreateShipmentRequest{
		ReferenceNumber:       "REF-100",
		Origin:                "A",
		Destination:           "B",
		DriverName:            "Driver",
		DriverPhone:           "Phone",
		UnitNumber:            "Unit",
		ShipmentAmount:        100.50,
		ShipmentCurrency:      "KZT",
		DriverRevenue:         80.00,
		DriverRevenueCurrency: "KZT",
	}

	shipment, err := uc.CreateShipment(context.Background(), req)
	require.NoError(t, err)

	assert.NotNil(t, shipment)
	assert.Equal(t, "REF-100", shipment.ReferenceNumber)
	assert.Equal(t, int64(10050), shipment.ShipmentAmount.Cents())
	assert.Equal(t, "KZT", shipment.ShipmentAmount.Currency())
}

func TestShipmentUseCase_UpdateStatus(t *testing.T) {
	svc := service.NewShipmentService(newMockShipmentRepo(), newMockStatusEventRepo(), &mockLogRepo{})
	uc := usecase.NewShipmentUseCase(svc)
	ctx := context.Background()

	// Seed first
	req := dto.CreateShipmentRequest{
		ReferenceNumber: "REF-100",
		Origin:          "A",
		Destination:     "B",
	}
	s, _ := uc.CreateShipment(ctx, req)
	time.Sleep(1 * time.Millisecond) // buffer for timestamp diff

	// Test update via DTO
	updateReq := dto.UpdateStatusRequest{
		ShipmentID: s.ID.String(),
		Status:     "picked_up",
		Location:   "Warehouse",
		Notes:      "Arrived",
	}

	event, err := uc.UpdateStatus(ctx, updateReq)
	require.NoError(t, err)

	assert.NotNil(t, event)
	assert.Equal(t, s.ID, event.ShipmentID)
	assert.Equal(t, valueobject.StatusPickedUp, event.Status)
}
