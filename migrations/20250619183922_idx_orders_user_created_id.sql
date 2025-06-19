-- +goose Up
CREATE INDEX idx_orders_user_created_id
    ON orders (user_id, created_at, id);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_user_created_id;
