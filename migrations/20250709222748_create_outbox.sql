-- +goose Up

CREATE TABLE outbox (
    id              UUID              PRIMARY KEY,
    payload         JSONB             NOT NULL,
    status          event_status      NOT NULL DEFAULT 'CREATED',
    error           TEXT,
    attempts        INT               NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ       NOT NULL DEFAULT now(),
    sent_at         TIMESTAMPTZ
);

CREATE INDEX idx_outbox_status     ON outbox(status);
CREATE INDEX idx_outbox_created_at ON outbox(created_at);

-- +goose Down
DROP TABLE IF EXISTS outbox;
