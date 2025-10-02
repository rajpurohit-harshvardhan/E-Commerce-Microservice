-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash STRING NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now()
    );

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
