package jsonfile

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/domain/entity"
)

type LogRepository struct {
	store *Store
}

func NewLogRepository(store *Store) *LogRepository {
	return &LogRepository{store: store}
}

func (r *LogRepository) Create(_ context.Context, l *entity.Log) error {
	return r.store.SaveLog(LogRecord{
		ID:        l.ID.String(),
		Action:    l.Action,
		Payload:   l.Payload,
		CreatedAt: TimeToString(l.CreatedAt),
	})
}

func (r *LogRepository) GetByAction(_ context.Context, action string) ([]*entity.Log, error) {
	records, err := r.store.GetLogsByAction(action)
	if err != nil {
		return nil, err
	}

	logs := make([]*entity.Log, 0, len(records))
	for _, rec := range records {
		l, err := recordToLog(&rec)
		if err != nil {
			return nil, fmt.Errorf("converting log record: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func recordToLog(r *LogRecord) (*entity.Log, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	createdAt, _ := StringToTime(r.CreatedAt)

	return &entity.Log{
		ID:        id,
		Action:    r.Action,
		Payload:   r.Payload,
		CreatedAt: createdAt,
	}, nil
}
