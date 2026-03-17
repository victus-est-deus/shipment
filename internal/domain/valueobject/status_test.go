package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victus-est-deus/shipment/internal/domain/valueobject"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{"pending", true},
		{"picked_up", true},
		{"in_transit", true},
		{"delivered", true},
		{"cancelled", true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			s := valueobject.Status(tt.status)
			assert.Equal(t, tt.expected, s.IsValid())
		})
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     valueobject.Status
		to       valueobject.Status
		expected bool
	}{
		// Valid transitions
		{"pending to picked_up", valueobject.StatusPending, valueobject.StatusPickedUp, true},
		{"pending to cancelled", valueobject.StatusPending, valueobject.StatusCancelled, true},
		{"picked_up to in_transit", valueobject.StatusPickedUp, valueobject.StatusInTransit, true},
		{"picked_up to cancelled", valueobject.StatusPickedUp, valueobject.StatusCancelled, true},
		{"in_transit to delivered", valueobject.StatusInTransit, valueobject.StatusDelivered, true},
		{"in_transit to cancelled", valueobject.StatusInTransit, valueobject.StatusCancelled, true},

		// Invalid transitions (skipping states)
		{"pending to in_transit", valueobject.StatusPending, valueobject.StatusInTransit, false},
		{"pending to delivered", valueobject.StatusPending, valueobject.StatusDelivered, false},
		{"picked_up to delivered", valueobject.StatusPickedUp, valueobject.StatusDelivered, false},

		// Invalid transitions (backwards)
		{"picked_up to pending", valueobject.StatusPickedUp, valueobject.StatusPending, false},
		{"in_transit to picked_up", valueobject.StatusInTransit, valueobject.StatusPickedUp, false},
		{"delivered to in_transit", valueobject.StatusDelivered, valueobject.StatusInTransit, false},

		// Invalid transitions (from terminal states)
		{"delivered to cancelled", valueobject.StatusDelivered, valueobject.StatusCancelled, false},
		{"cancelled to pending", valueobject.StatusCancelled, valueobject.StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.from.CanTransitionTo(tt.to))
		})
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	assert.False(t, valueobject.StatusPending.IsTerminal())
	assert.False(t, valueobject.StatusPickedUp.IsTerminal())
	assert.False(t, valueobject.StatusInTransit.IsTerminal())
	assert.True(t, valueobject.StatusDelivered.IsTerminal())
	assert.True(t, valueobject.StatusCancelled.IsTerminal())
}

func TestParseStatus(t *testing.T) {
	s, err := valueobject.ParseStatus("pending")
	require.NoError(t, err)
	assert.Equal(t, valueobject.StatusPending, s)

	_, err = valueobject.ParseStatus("invalid_status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid shipment status")
}
