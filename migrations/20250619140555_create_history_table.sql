-- +goose Up
CREATE TABLE order_history (
    id         BIGSERIAL     PRIMARY KEY,
    order_id   BIGINT          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    status     order_status  NOT NULL,
    event_time TIMESTAMPTZ   NOT NULL
);

-- +goose Down
DROP TABLE  IF EXISTS order_history;