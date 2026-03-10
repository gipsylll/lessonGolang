-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL      PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL,
    version    INTEGER     NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Уникальность email + быстрый поиск по нему
CREATE UNIQUE INDEX IF NOT EXISTS users_email_uidx ON users (email);

-- Для сортировки по дате создания (пагинация в будущем)
CREATE INDEX IF NOT EXISTS users_created_at_idx ON users (created_at);

-- +goose Down
DROP TABLE IF EXISTS users;
