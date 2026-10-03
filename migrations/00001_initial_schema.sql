-- +goose Up

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    display_name TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_email_unique
    ON users (LOWER(email));

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hashed_key TEXT NOT NULL UNIQUE,
    encrypted_key TEXT,
    internal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX api_keys_user_id_idx
    ON api_keys (user_id);

CREATE INDEX api_keys_active_user_id_idx
    ON api_keys (user_id)
    WHERE revoked_at IS NULL;


-- +goose Down

DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;