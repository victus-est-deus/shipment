package entity

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID        uuid.UUID
	Action    string
	Payload   map[string]any
	CreatedAt time.Time
}

func NewLog(action string, payload map[string]any) *Log {
	if payload == nil {
		payload = make(map[string]any)
	}

	return &Log{
		ID:        uuid.New(),
		Action:    action,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}
}
