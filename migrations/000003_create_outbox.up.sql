CREATE TABLE IF NOT EXISTS outbox_messages (
    id           BIGSERIAL PRIMARY KEY,
    topic        TEXT        NOT NULL,
    payload      TEXT        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox_messages (status) WHERE status = 'pending';
