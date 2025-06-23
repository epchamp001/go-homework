-- +goose Up
CREATE INDEX idx_history_event_order
    ON order_history (event_time DESC, order_id DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_history_event_order;
