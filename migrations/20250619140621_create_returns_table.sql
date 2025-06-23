-- +goose Up
CREATE TABLE order_returns (
    order_id     BIGINT        PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    user_id      BIGINT       NOT NULL,
    returned_at  TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE  IF EXISTS order_returns;