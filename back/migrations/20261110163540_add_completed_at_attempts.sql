-- +goose Up

ALTER TABLE attempts
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE attempts
    DROP COLUMN IF EXISTS completed_at;