-- +goose Up

ALTER TABLE users
    ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

ALTER TABLE users
    ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'user'));

-- +goose Down

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
    DROP COLUMN IF EXISTS role;
