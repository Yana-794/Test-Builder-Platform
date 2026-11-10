-- +goose Up
-- +goose StatementBegin

CREATE TABLE test_publications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id      UUID NOT NULL REFERENCES tests(id) ON DELETE CASCADE,
    version      INT  NOT NULL,
    public_token TEXT NOT NULL UNIQUE,
    content      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (test_id, version)
);

CREATE INDEX idx_test_publications_test_id ON test_publications(test_id);

ALTER TABLE tests
    ADD COLUMN IF NOT EXISTS publication_version_id UUID
        REFERENCES test_publications(id) ON DELETE SET NULL;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

ALTER TABLE tests DROP COLUMN IF EXISTS publication_version_id;
DROP TABLE IF EXISTS test_publications;

-- +goose StatementEnd