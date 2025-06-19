-- +goose Up
CREATE TYPE order_status   AS ENUM ('ACCEPTED', 'ISSUED', 'RETURNED', 'EXPIRED');
CREATE TYPE package_type   AS ENUM ('', 'bag', 'box', 'film', 'bag+film', 'box+film');

-- +goose Down
DROP TYPE package_type;
DROP TYPE order_status;

