-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name STRING,
    email STRING UNIQUE NOT NULL,
    password_hash STRING NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS users;
