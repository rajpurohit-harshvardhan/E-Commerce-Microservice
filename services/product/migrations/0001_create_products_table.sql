-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku STRING UNIQUE NOT NULL,
    name STRING NOT NULL,
    description STRING,
    price DECIMAL(10,2) NOT NULL,
    stock INT8 NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS products;