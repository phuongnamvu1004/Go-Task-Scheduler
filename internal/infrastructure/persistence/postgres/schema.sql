-- Runs on every startup, so statements must be idempotent.
-- TODO: replace with a migration tool (e.g. goose).

CREATE TABLE IF NOT EXISTS jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    schedule    TEXT        NOT NULL,
    status      TEXT        NOT NULL CHECK (status IN ('active', 'paused')),
    task_type   TEXT        NOT NULL,
    task_config JSONB       NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS jobs_task_type_idx ON jobs (task_type);
