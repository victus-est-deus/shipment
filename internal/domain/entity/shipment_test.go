package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

func TestNewShipment(t *testing.T) {
	amount, _ := valueobject.NewMoney(100.50, "USD")
	revenue, _ := valueobject.NewMoney(80.00, "USD")

	t.Run("success", func(t *testing.T) {
		s, err := entity.NewShipment(
			"REF-123", "Almaty", "Astana",
			"John Doe", "+123", "TRK-1",
			amount, revenue,
		)

		require.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotEmpty(t, s.ID)
		assert.Equal(t, "REF-123", s.ReferenceNumber)
		assert.Equal(t, valueobject.StatusPending, s.Status)
		assert.WithinDuration(t, time.Now().UTC(), s.CreatedAt, 2*time.Second)
		assert.Equal(t, s.CreatedAt, s.UpdatedAt)
	})

	t.Run("missing reference number", func(t *testing.T) {
		_, err := entity.NewShipment("", "Almaty", "Astana", "John", "+1", "T1", amount, revenue)
		assert.ErrorContains(t, err, "reference number must not be empty")
	})

	t.Run("missing origin", func(t *testing.T) {
		_, err := entity.NewShipment("REF", "", "Astana", "John", "+1", "T1", amount, revenue)
		assert.ErrorContains(t, err, "origin must not be empty")
	})

	t.Run("missing destination", func(t *testing.T) {
		_, err := entity.NewShipment("REF", "Almaty", "", "John", "+1", "T1", amount, revenue)
		assert.ErrorContains(t, err, "destination must not be empty")
	})
}

func TestShipment_AddStatusEvent(t *testing.T) {
	amount, _ := valueobject.NewMoney(100.0, "USD")
	revenue, _ := valueobject.NewMoney(80.0, "USD")

	s, err := entity.NewShipment("REF-123", "Orig", "Dest", "Driver", "Phone", "Unit", amount, revenue)
	require.NoError(t, err)

	// Ensure initial state
	assert.Equal(t, valueobject.StatusPending, s.Status)

	t.Run("valid transition", func(t *testing.T) {
		oldUpdatedAt := s.UpdatedAt

		// Add small delay to ensure UpdatedAt changes
		time.Sleep(1 * time.Millisecond)

		event, err := s.AddStatusEvent(valueobject.StatusPickedUp, "Warehouse A", "Driver arrived")
		require.NoError(t, err)

		// Check event
		assert.NotNil(t, event)
		assert.NotEmpty(t, event.ID)
		assert.Equal(t, s.ID, event.ShipmentID)
		assert.Equal(t, valueobject.StatusPickedUp, event.Status)
		assert.Equal(t, "Warehouse A", event.Location)
		assert.Equal(t, "Driver arrived", event.Notes)

		// Check shipment mutation
		assert.Equal(t, valueobject.StatusPickedUp, s.Status)
		assert.True(t, s.UpdatedAt.After(oldUpdatedAt))
		assert.Equal(t, event.CreatedAt, s.UpdatedAt)
	})

	t.Run("invalid transition", func(t *testing.T) {
		// Currently 'picked_up', cannot jump to 'delivered'
		_, err := s.AddStatusEvent(valueobject.StatusDelivered, "", "")
		assert.ErrorContains(t, err, "invalid status transition")
		assert.Equal(t, valueobject.StatusPickedUp, s.Status) // state unchanged
	})

	t.Run("invalid status string", func(t *testing.T) {
		_, err := s.AddStatusEvent(valueobject.Status("unknown"), "", "")
		assert.ErrorContains(t, err, "invalid status")
	})
}
