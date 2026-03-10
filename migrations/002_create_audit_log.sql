-- +goose Up
CREATE TABLE IF NOT EXISTS audit_log (
    id         BIGSERIAL    PRIMARY KEY,
    entity_id  INTEGER      NOT NULL,
    action     TEXT         NOT NULL, -- create | update | patch
    old_data   JSONB,                 -- NULL для create
    new_data   JSONB        NOT NULL,
    changed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Поиск всех изменений конкретного пользователя по времени (DESC — свежее сначала)
CREATE INDEX IF NOT EXISTS audit_log_entity_id_idx ON audit_log (entity_id, changed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_log;
