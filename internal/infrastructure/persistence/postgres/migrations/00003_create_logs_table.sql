-- +goose Up
CREATE TABLE logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    action      VARCHAR(100) NOT NULL,
    payload     JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_logs_action ON logs (action);
CREATE INDEX idx_logs_created_at ON logs (created_at);

-- +goose Down
DROP TABLE IF EXISTS logs;
