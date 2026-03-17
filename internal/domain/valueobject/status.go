package valueobject

import "fmt"

type Status string

const (
	StatusPending   Status = "pending"
	StatusPickedUp  Status = "picked_up"
	StatusInTransit Status = "in_transit"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

var validTransitions = map[Status]map[Status]bool{
	StatusPending:   {StatusPickedUp: true, StatusCancelled: true},
	StatusPickedUp:  {StatusInTransit: true, StatusCancelled: true},
	StatusInTransit: {StatusDelivered: true, StatusCancelled: true},
	StatusDelivered: {},
	StatusCancelled: {},
}

func AllStatuses() []Status {
	return []Status{
		StatusPending,
		StatusPickedUp,
		StatusInTransit,
		StatusDelivered,
		StatusCancelled,
	}
}

func (s Status) IsValid() bool {
	_, exists := validTransitions[s]
	return exists
}

func (s Status) CanTransitionTo(target Status) bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	return allowed[target]
}

func (s Status) IsTerminal() bool {
	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}
	return len(allowed) == 0
}

func (s Status) String() string {
	return string(s)
}

func ParseStatus(s string) (Status, error) {
	status := Status(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid shipment status: %q", s)
	}
	return status, nil
}
