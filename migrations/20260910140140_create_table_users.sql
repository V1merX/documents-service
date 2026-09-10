-- +goose Up
CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    token TEXT UNIQUE DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_token ON users(token) WHERE token IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS users;