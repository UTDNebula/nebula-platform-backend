-- +goose Up

CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider TEXT NOT NULL,
    issuer TEXT NOT NULL,
    subject TEXT NOT NULL,

    email TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT user_identities_issuer_subject_unique UNIQUE (issuer, subject)
);

CREATE INDEX user_identities_user_id_idx
    ON user_identities (user_id);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx
    ON sessions (expiry);


-- +goose Down

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_identities;