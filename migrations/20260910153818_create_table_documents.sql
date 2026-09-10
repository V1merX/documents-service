-- +goose Up
CREATE TABLE IF NOT EXISTS documents
(
    id           UUID PRIMARY KEY,
    owner        TEXT        NOT NULL REFERENCES users(login),
    name         TEXT        NOT NULL,
    mime         TEXT        NOT NULL DEFAULT '',
    file         BOOLEAN     NOT NULL,
    public       BOOLEAN     NOT NULL DEFAULT false,
    json_data    JSONB,
    content_path TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    shared_with  TEXT[]      NOT NULL DEFAULT '{}',
    CONSTRAINT chk_file_has_path CHECK (
        (
            file = true
                AND content_path IS NOT NULL
            )
            OR (
            file = false
                AND content_path IS NULL
            )
        )
);
CREATE INDEX IF NOT EXISTS idx_documents_owner ON documents (owner);
CREATE INDEX IF NOT EXISTS idx_documents_created ON documents (name, created_at);
CREATE INDEX IF NOT EXISTS idx_documents_shared ON documents USING GIN (shared_with);

-- +goose Down
DROP TABLE IF EXISTS documents;
