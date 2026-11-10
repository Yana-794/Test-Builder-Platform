-- +goose Up
-- +goose StatementBegin

CREATE TABLE attempts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publication_id    UUID NOT NULL REFERENCES test_publications(id) ON DELETE CASCADE,
    student_name      TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'in_progress'
                          CHECK (status IN ('in_progress', 'completed')),
    access_token_hash BYTEA NOT NULL,
    started_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ NOT NULL,   
    score             NUMERIC(10, 2),
    max_score         NUMERIC(10, 2),
    score_percent     NUMERIC(5, 2),          
    duration_seconds  INT
);

CREATE INDEX idx_attempts_publication_id ON attempts(publication_id);

CREATE INDEX idx_attempts_expires_at_in_progress ON attempts(expires_at)
    WHERE status = 'in_progress';

CREATE TABLE attempt_answers (
    attempt_id         UUID NOT NULL REFERENCES attempts(id) ON DELETE CASCADE,
    question_id        UUID NOT NULL,
    answer_text        TEXT,
    selected_option_id UUID,

    PRIMARY KEY (attempt_id, question_id)
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS attempt_answers;
DROP TABLE IF EXISTS attempts;

-- +goose StatementEnd