CREATE TABLE IF NOT EXISTS viewings (
    id               BIGSERIAL PRIMARY KEY,
    agent_id         BIGINT      NOT NULL,
    lead_id          BIGINT      NOT NULL,
    property_address TEXT        NOT NULL,
    scheduled_at     TIMESTAMPTZ NOT NULL,
    status           TEXT        NOT NULL DEFAULT 'SCHEDULED',
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ix_viewings_agent_scheduled_at
    ON viewings (agent_id, scheduled_at);

CREATE INDEX IF NOT EXISTS ix_viewings_id_scheduled_at
    ON viewings (scheduled_at, id);

CREATE INDEX IF NOT EXISTS ix_viewings_status
    ON viewings (status);

CREATE INDEX IF NOT EXISTS ix_viewings_status_scheduled_at
    ON viewings (status, scheduled_at);