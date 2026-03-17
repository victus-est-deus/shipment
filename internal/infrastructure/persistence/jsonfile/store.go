package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/victus-est-deus/shipment/internal/infrastructure/config"
)

type ShipmentRecord struct {
	ID              string  `json:"id"`
	ReferenceNumber string  `json:"reference_number"`
	Origin          string  `json:"origin"`
	Destination     string  `json:"destination"`
	Status          string  `json:"status"`
	DriverName      string  `json:"driver_name"`
	DriverPhone     string  `json:"driver_phone"`
	UnitNumber      string  `json:"unit_number"`
	ShipmentAmount  int64   `json:"shipment_amount"`
	DriverRevenue   int64   `json:"driver_revenue"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type LogRecord struct {
	ID        string         `json:"id"`
	Action    string         `json:"action"`
	Payload   map[string]any `json:"payload"`
	CreatedAt string         `json:"created_at"`
}

type StatusEventRecord struct {
	ID         string `json:"id"`
	ShipmentID string `json:"shipment_id"`
	Status     string `json:"status"`
	Location   string `json:"location"`
	Notes      string `json:"notes"`
	CreatedAt  string `json:"created_at"`
}

type Store struct {
	basePath string
	mu       sync.RWMutex
}

func NewStore(basePath string) (*Store, error) {
	dirs := []string{
		filepath.Join(basePath, "shipments"),
		filepath.Join(basePath, "status_events"),
		filepath.Join(basePath, "logs"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return &Store{basePath: basePath}, nil
}

func NewStoreFromConfig() (*Store, error) {
	return NewStore(config.DefaultStoragePath)
}

func (s *Store) SaveShipment(record ShipmentRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.writeJSON(filepath.Join(s.basePath, "shipments", record.ID+".json"), record)
}

func (s *Store) GetShipment(id uuid.UUID) (*ShipmentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var record ShipmentRecord
	err := s.readJSON(filepath.Join(s.basePath, "shipments", id.String()+".json"), &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *Store) SaveStatusEvent(record StatusEventRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.writeJSON(filepath.Join(s.basePath, "status_events", record.ID+".json"), record)
}

func (s *Store) GetStatusEventsByShipmentID(shipmentID uuid.UUID) ([]StatusEventRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.basePath, "status_events")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading status_events directory: %w", err)
	}

	var events []StatusEventRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var record StatusEventRecord
		if err := s.readJSON(filepath.Join(dir, entry.Name()), &record); err != nil {
			continue
		}

		if record.ShipmentID == shipmentID.String() {
			events = append(events, record)
		}
	}

	return events, nil
}

func (s *Store) SaveLog(record LogRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.writeJSON(filepath.Join(s.basePath, "logs", record.ID+".json"), record)
}

func (s *Store) GetLogsByAction(action string) ([]LogRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.basePath, "logs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading logs directory: %w", err)
	}

	var logs []LogRecord
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var record LogRecord
		if err := s.readJSON(filepath.Join(dir, entry.Name()), &record); err != nil {
			continue
		}

		if action == "" || record.Action == action {
			logs = append(logs, record)
		}
	}

	return logs, nil
}

func (s *Store) writeJSON(path string, data any) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", path, err)
	}
	return nil
}

func (s *Store) readJSON(path string, dest any) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("unmarshalling JSON: %w", err)
	}
	return nil
}

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func StringToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
