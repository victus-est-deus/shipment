package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/victus-est-deus/shipment/internal/domain/entity"
)

type LogRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Create(ctx context.Context, l *entity.Log) error {
	payload, err := json.Marshal(l.Payload)
	if err != nil {
		return fmt.Errorf("marshalling log payload: %w", err)
	}

	query := `
		INSERT INTO logs (id, action, payload, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err = r.db.ExecContext(ctx, query, l.ID, l.Action, payload, l.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting log: %w", err)
	}
	return nil
}

func (r *LogRepository) GetByAction(ctx context.Context, action string) ([]*entity.Log, error) {
	query := `SELECT id, action, payload, created_at FROM logs WHERE action = $1 ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, action)
	if err != nil {
		return nil, fmt.Errorf("querying logs: %w", err)
	}
	defer rows.Close()

	var logs []*entity.Log
	for rows.Next() {
		var l entity.Log
		var payload []byte

		if err := rows.Scan(&l.ID, &l.Action, &payload, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning log: %w", err)
		}

		if err := json.Unmarshal(payload, &l.Payload); err != nil {
			return nil, fmt.Errorf("unmarshalling log payload: %w", err)
		}

		logs = append(logs, &l)
	}

	return logs, rows.Err()
}
