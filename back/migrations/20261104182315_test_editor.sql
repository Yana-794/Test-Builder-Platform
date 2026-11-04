-- +goose Up

ALTER TABLE tests
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    ADD COLUMN subject     VARCHAR(100),
    ADD COLUMN revision    INTEGER NOT NULL DEFAULT 1;

ALTER TABLE tests DROP CONSTRAINT IF EXISTS tests_title_check;
ALTER TABLE tests
    ADD CONSTRAINT tests_title_check
    CHECK (char_length(btrim(title)) BETWEEN 1 AND 160);

CREATE INDEX idx_tests_owner_updated ON tests (owner_id, updated_at DESC);

CREATE TABLE questions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id             UUID NOT NULL REFERENCES tests (id) ON DELETE CASCADE,
    prompt              TEXT NOT NULL,
    question_type       VARCHAR(30) NOT NULL
                            CHECK (question_type IN ('single_choice', 'text')),
    points              NUMERIC(9, 2) NOT NULL
                            CHECK (points > 0),
    position            INTEGER NOT NULL
                            CHECK (position >= 0),
    correct_text_answer TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT questions_correct_answer_by_type
        CHECK (question_type <> 'single_choice' OR correct_text_answer IS NULL),

    CONSTRAINT uq_questions_test_position UNIQUE (test_id, position)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE TRIGGER trg_questions_set_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE answer_options (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id  UUID NOT NULL REFERENCES questions (id) ON DELETE CASCADE,
    option_text  VARCHAR(500) NOT NULL,
    is_correct   BOOLEAN NOT NULL DEFAULT false,
    position     INTEGER NOT NULL CHECK (position >= 0),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_answer_options_question_position UNIQUE (question_id, position)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE UNIQUE INDEX uq_answer_options_one_correct
    ON answer_options (question_id)
    WHERE is_correct = true;

CREATE TRIGGER trg_answer_options_set_updated_at
    BEFORE UPDATE ON answer_options
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TABLE IF EXISTS answer_options;
DROP TABLE IF EXISTS questions;

DROP INDEX IF EXISTS idx_tests_owner_updated;

ALTER TABLE tests DROP CONSTRAINT IF EXISTS tests_title_check;
ALTER TABLE tests
    ADD CONSTRAINT tests_title_check
    CHECK (char_length(btrim(title)) BETWEEN 1 AND 200);

ALTER TABLE tests
    DROP COLUMN IF EXISTS revision,
    DROP COLUMN IF EXISTS subject,
    DROP COLUMN IF EXISTS description;