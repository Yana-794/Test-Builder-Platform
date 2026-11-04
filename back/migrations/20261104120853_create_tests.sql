-- +goose Up
CREATE TABLE tests (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    uuid        NOT NULL
                            REFERENCES users (id)
                            ON DELETE CASCADE,
    title       text        NOT NULL
                            CHECK (char_length(btrim(title)) BETWEEN 1 AND 200),
    status      text        NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft', 'published', 'archived')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_tests_owner_active_created
    ON tests (owner_id, created_at DESC, id DESC)
    WHERE status IN ('draft', 'published');

CREATE INDEX idx_tests_owner_status
    ON tests (owner_id, status);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_tests_set_updated_at
    BEFORE UPDATE ON tests
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE IF EXISTS tests;