-- +goose Up
CREATE TABLE orders (
    id           BIGSERIAL        PRIMARY KEY,
    user_id      BIGINT        NOT NULL,
    status       order_status   NOT NULL,
    expires_at   TIMESTAMPTZ    NOT NULL,
    issued_at    TIMESTAMPTZ,
    returned_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT now(),
    package      package_type   NOT NULL,
    weight       NUMERIC(10,3)  NOT NULL,
    price        BIGINT         NOT NULL,
    total_price  BIGINT         NOT NULL
);

-- +goose Down
DROP TABLE  IF EXISTS orders;