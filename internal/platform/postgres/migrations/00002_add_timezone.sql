-- The default fills in existing rows; the app always sends a value.

-- +goose Up
ALTER TABLE jobs ADD COLUMN timezone TEXT NOT NULL DEFAULT 'UTC';

-- +goose Down
ALTER TABLE jobs DROP COLUMN timezone;
