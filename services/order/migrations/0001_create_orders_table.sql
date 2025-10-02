-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status STRING NOT NULL DEFAULT 'pending',
    total DECIMAL NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS orders;
