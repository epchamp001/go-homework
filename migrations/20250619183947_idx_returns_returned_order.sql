-- +goose Up
CREATE INDEX idx_returns_returned_order
    ON order_returns (returned_at DESC, order_id DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_returns_returned_order;
