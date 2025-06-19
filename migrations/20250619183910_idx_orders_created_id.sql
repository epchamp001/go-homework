-- +goose Up
CREATE INDEX idx_orders_created_id
    ON orders (created_at, id);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_created_id;
