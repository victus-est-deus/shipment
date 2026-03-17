package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/service"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

// --- Mocks ---

type mockShipmentRepo struct {
	shipments map[uuid.UUID]*entity.Shipment
	byRef     map[string]*entity.Shipment
}

func newMockShipmentRepo() *mockShipmentRepo {
	return &mockShipmentRepo{
		shipments: make(map[uuid.UUID]*entity.Shipment),
		byRef:     make(map[string]*entity.Shipment),
	}
}

func (m *mockShipmentRepo) Create(ctx context.Context, s *entity.Shipment) error {
	m.shipments[s.ID] = s
	m.byRef[s.ReferenceNumber] = s
	return nil
}

func (m *mockShipmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Shipment, error) {
	if s, ok := m.shipments[id]; ok {
		return s, nil
	}
	return nil, service.ErrShipmentNotFound
}

func (m *mockShipmentRepo) GetByReferenceNumber(ctx context.Context, ref string) (*entity.Shipment, error) {
	if s, ok := m.byRef[ref]; ok {
		return s, nil
	}
	return nil, service.ErrShipmentNotFound
}

func (m *mockShipmentRepo) Update(ctx context.Context, s *entity.Shipment) error {
	m.shipments[s.ID] = s
	m.byRef[s.ReferenceNumber] = s
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

type mockLogRepo struct {
	logs []*entity.Log
}

func (m *mockLogRepo) Create(ctx context.Context, l *entity.Log) error {
	m.logs = append(m.logs, l)
	return nil
}

func (m *mockLogRepo) GetByAction(ctx context.Context, action string) ([]*entity.Log, error) {
	var filtered []*entity.Log
	for _, l := range m.logs {
		if l.Action == action {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

// --- Tests ---

func TestShipmentService_CreateShipment(t *testing.T) {
	shipmentRepo := newMockShipmentRepo()
	eventRepo := newMockStatusEventRepo()
	logRepo := &mockLogRepo{}

	svc := service.NewShipmentService(shipmentRepo, eventRepo, logRepo)

	amount, _ := valueobject.NewMoney(100.0, "USD")
	revenue, _ := valueobject.NewMoney(80.0, "USD")

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		s, err := svc.CreateShipment(ctx, "REF-001", "Almaty", "Astana", "Driver 1", "Phone 1", "Unit 1", amount, revenue)
		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, valueobject.StatusPending, s.Status)

		// Check persistence
		stored, err := shipmentRepo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, s.ID, stored.ID)

		// Check status event was created
		events, _ := eventRepo.GetByShipmentID(ctx, s.ID)
		require.Len(t, events, 1)
		assert.Equal(t, valueobject.StatusPending, events[0].Status)
		assert.Equal(t, "shipment created", events[0].Notes)

		// Check audit log
		assert.Len(t, logRepo.logs, 1)
		assert.Equal(t, "shipment_created", logRepo.logs[0].Action)
	})

	t.Run("duplicate reference number", func(t *testing.T) {
		_, err := svc.CreateShipment(ctx, "REF-001", "A", "B", "C", "D", "E", amount, revenue)
		assert.ErrorIs(t, err, service.ErrDuplicateReference)
	})
}

func TestShipmentService_UpdateStatus(t *testing.T) {
	shipmentRepo := newMockShipmentRepo()
	eventRepo := newMockStatusEventRepo()
	logRepo := &mockLogRepo{}

	svc := service.NewShipmentService(shipmentRepo, eventRepo, logRepo)
	amount, _ := valueobject.NewMoney(100.0, "USD")
	revenue, _ := valueobject.NewMoney(80.0, "USD")
	ctx := context.Background()

	// Seed a shipment
	s, _ := svc.CreateShipment(ctx, "REF-002", "Almaty", "Astana", "Driver 2", "Phone 2", "Unit 2", amount, revenue)

	t.Run("valid transition: pending -> picked_up", func(t *testing.T) {
		// allow time tick for updated_at
		time.Sleep(1 * time.Millisecond)

		event, err := svc.UpdateStatus(ctx, s.ID, valueobject.StatusPickedUp, "Warehouse", "Arrived")
		require.NoError(t, err)
		assert.Equal(t, valueobject.StatusPickedUp, event.Status)

		stored, _ := shipmentRepo.GetByID(ctx, s.ID)
		assert.Equal(t, valueobject.StatusPickedUp, stored.Status)

		events, _ := eventRepo.GetByShipmentID(ctx, s.ID)
		assert.Len(t, events, 2) // initial + picked_up

		// Audit logs: created + updated
		assert.Len(t, logRepo.logs, 2)
		assert.Equal(t, "status_updated", logRepo.logs[1].Action)
	})

	t.Run("invalid transition: picked_up -> delivered", func(t *testing.T) {
		_, err := svc.UpdateStatus(ctx, s.ID, valueobject.StatusDelivered, "", "")
		assert.ErrorIs(t, err, service.ErrInvalidTransition)

		stored, _ := shipmentRepo.GetByID(ctx, s.ID)
		assert.Equal(t, valueobject.StatusPickedUp, stored.Status) // remains picked_up
	})

	t.Run("transition to terminal state", func(t *testing.T) {
		_, err := svc.UpdateStatus(ctx, s.ID, valueobject.StatusInTransit, "", "")
		require.NoError(t, err)

		_, err = svc.UpdateStatus(ctx, s.ID, valueobject.StatusDelivered, "", "")
		require.NoError(t, err)

		stored, _ := shipmentRepo.GetByID(ctx, s.ID)
		assert.Equal(t, valueobject.StatusDelivered, stored.Status)
	})

	t.Run("update terminal state", func(t *testing.T) {
		// Shipment REF-002 is now delivered.
		_, err := svc.UpdateStatus(ctx, s.ID, valueobject.StatusCancelled, "", "")
		assert.ErrorIs(t, err, service.ErrShipmentTerminated)

		stored, _ := shipmentRepo.GetByID(ctx, s.ID)
		assert.Equal(t, valueobject.StatusDelivered, stored.Status)
	})

	t.Run("shipment not found", func(t *testing.T) {
		_, err := svc.UpdateStatus(ctx, uuid.New(), valueobject.StatusPickedUp, "", "")
		assert.ErrorIs(t, err, service.ErrShipmentNotFound)
	})
}
