-- +goose Up

CREATE TYPE event_status AS ENUM (
    'CREATED',
    'PROCESSING',
    'COMPLETED',
    'FAILED'
);

-- +goose Down
DROP TYPE IF EXISTS event_status;
