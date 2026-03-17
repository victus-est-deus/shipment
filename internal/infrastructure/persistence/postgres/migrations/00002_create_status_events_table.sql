-- +goose Up
CREATE TABLE status_events (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shipment_id UUID         NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status      VARCHAR(50)  NOT NULL,
    location    VARCHAR(255) NOT NULL DEFAULT '',
    notes       TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_events_shipment_id ON status_events (shipment_id);
CREATE INDEX idx_status_events_created_at ON status_events (shipment_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS status_events;
