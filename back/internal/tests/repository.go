package tests

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// общий интерфейс 
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateTest

func (r *Repository) CreateTest(
	ctx context.Context,
	ownerID uuid.UUID,
	title, description string,
	subject *string,
) (Test, error) {
	const q = `
		INSERT INTO tests (owner_id, title, description, subject, status, revision)
		VALUES ($1, $2, $3, $4, 'draft', 1)
		RETURNING id, owner_id, title, description, subject,
		          status, revision, created_at, updated_at
	`
	var t Test
	err := r.pool.QueryRow(ctx, q, ownerID, title, description, subject).Scan(
		&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Subject,
		&t.Status, &t.Revision, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return Test{}, fmt.Errorf("create test: %w", err)
	}
	return t, nil
}

// GetTest 

func (r *Repository) GetTest(ctx context.Context, ownerID, testID uuid.UUID) (Test, error) {
	return r.getTest(ctx, r.pool, ownerID, testID)
}

func (r *Repository) getTest(ctx context.Context, q querier, ownerID, testID uuid.UUID) (Test, error) {
	const query = `
		SELECT id, owner_id, title, description, subject,
		       status, revision, created_at, updated_at
		FROM tests
		WHERE id = $1 AND owner_id = $2
	`
	var t Test
	err := q.QueryRow(ctx, query, testID, ownerID).Scan(
		&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Subject,
		&t.Status, &t.Revision, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Test{}, ErrTestNotFound
	}
	if err != nil {
		return Test{}, fmt.Errorf("get test: %w", err)
	}
	return t, nil
}

func (r *Repository) GetTestWithQuestions(
	ctx context.Context,
	ownerID, testID uuid.UUID,
) (Test, []Question, error) {
	return r.getTestWithQuestions(ctx, r.pool, ownerID, testID)
}

func (r *Repository) getTestWithQuestions(
	ctx context.Context,
	q querier,
	ownerID, testID uuid.UUID,
) (Test, []Question, error) {
	test, err := r.getTest(ctx, q, ownerID, testID)
	if err != nil {
		return Test{}, nil, err
	}
	questions, err := loadQuestions(ctx, q, testID)
	if err != nil {
		return Test{}, nil, err
	}
	return test, questions, nil
}

func loadQuestions(ctx context.Context, q querier, testID uuid.UUID) ([]Question, error) {
	const qry = `
		SELECT id, test_id, prompt, question_type,
		       points::float8, position, correct_text_answer
		FROM questions
		WHERE test_id = $1
		ORDER BY position
	`
	rows, err := q.Query(ctx, qry, testID)
	if err != nil {
		return nil, fmt.Errorf("query questions: %w", err)
	}
	defer rows.Close()

	questions := make([]Question, 0)
	questionIndex := make(map[uuid.UUID]int)

	for rows.Next() {
		var qn Question
		if err := rows.Scan(
			&qn.ID, &qn.TestID, &qn.Prompt, &qn.Type,
			&qn.Points, &qn.Position, &qn.CorrectTextAnswer,
		); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		qn.Options = make([]AnswerOption, 0)
		questionIndex[qn.ID] = len(questions)
		questions = append(questions, qn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows questions: %w", err)
	}
	if len(questions) == 0 {
		return questions, nil
	}

	optionQry := `
		SELECT id, question_id, option_text, is_correct, position
		FROM answer_options
		WHERE question_id = ANY($1)
		ORDER BY question_id, position
	`
	ids := make([]uuid.UUID, 0, len(questions))
	for _, qn := range questions {
		ids = append(ids, qn.ID)
	}

	optRows, err := q.Query(ctx, optionQry, ids)
	if err != nil {
		return nil, fmt.Errorf("query options: %w", err)
	}
	defer optRows.Close()

	for optRows.Next() {
		var opt AnswerOption
		if err := optRows.Scan(
			&opt.ID, &opt.QuestionID, &opt.Text, &opt.IsCorrect, &opt.Position,
		); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}
		if idx, ok := questionIndex[opt.QuestionID]; ok {
			questions[idx].Options = append(questions[idx].Options, opt)
		}
	}
	if err := optRows.Err(); err != nil {
		return nil, fmt.Errorf("rows options: %w", err)
	}

	return questions, nil
}

// SaveDraft — атомарное сохранение всего черновика

func (r *Repository) SaveDraft(
	ctx context.Context,
	ownerID, testID uuid.UUID,
	expectedRevision int,
	req SaveDraftRequest,
) (Test, []Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Test{}, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	
	var currentRevision int
	var currentStatus string
	err = tx.QueryRow(ctx, `
		SELECT revision, status
		FROM tests
		WHERE id = $1 AND owner_id = $2
		FOR UPDATE
	`, testID, ownerID).Scan(&currentRevision, &currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return Test{}, nil, ErrTestNotFound
	}
	if err != nil {
		return Test{}, nil, fmt.Errorf("lock test: %w", err)
	}
	if currentStatus != StatusDraft {
		return Test{}, nil, ErrTestNotDraft
	}
	if currentRevision != expectedRevision {
		return Test{}, nil, ErrRevisionConflict
	}
	_, err = tx.Exec(ctx, `
		UPDATE tests
		SET title = $1, description = $2, subject = $3,
		    revision = revision + 1
		WHERE id = $4 AND owner_id = $5
	`, req.Test.Title, req.Test.Description, req.Test.Subject, testID, ownerID)
	if err != nil {
		return Test{}, nil, fmt.Errorf("update test: %w", err)
	}

	keptQuestionIDs := make([]uuid.UUID, 0, len(req.Questions))
	for _, item := range req.Questions {
		if item.ID == nil {
			continue
		}
		parsed, perr := uuid.Parse(*item.ID)
		if perr != nil {
			return Test{}, nil, fmt.Errorf("%w: invalid question id %q", ErrInvalidReference, *item.ID)
		}
		keptQuestionIDs = append(keptQuestionIDs, parsed)
	}
	_, err = tx.Exec(ctx, `
		DELETE FROM questions
		WHERE test_id = $1
		  AND (cardinality($2::uuid[]) = 0 OR id <> ALL($2::uuid[]))
	`, testID, keptQuestionIDs)
	if err != nil {
		return Test{}, nil, fmt.Errorf("delete removed questions: %w", err)
	}

	for pos, item := range req.Questions {
		questionID, err := upsertQuestion(ctx, tx, testID, pos, item)
		if err != nil {
			return Test{}, nil, err
		}
		if err := upsertOptions(ctx, tx, questionID, item.Options); err != nil {
			return Test{}, nil, err
		}
	}

	test, questions, err := r.getTestWithQuestions(ctx, tx, ownerID, testID)
	if err != nil {
		return Test{}, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Test{}, nil, fmt.Errorf("commit tx: %w", err)
	}

	return test, questions, nil
}

func upsertQuestion(
	ctx context.Context,
	tx pgx.Tx,
	testID uuid.UUID,
	position int,
	item DraftItem,
) (uuid.UUID, error) {
	if item.ID != nil {
		id, _ := uuid.Parse(*item.ID) // валидность проверена до этого
		res, err := tx.Exec(ctx, `
			UPDATE questions
			SET prompt = $1, question_type = $2, points = $3,
			    position = $4, correct_text_answer = $5
			WHERE id = $6 AND test_id = $7
		`, item.Prompt, item.Type, item.Points, position,
			item.CorrectTextAnswer, id, testID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("update question: %w", err)
		}
		if res.RowsAffected() == 0 {
			return uuid.Nil, fmt.Errorf(
				"%w: question %s does not belong to this test",
				ErrInvalidReference, id,
			)
		}
		return id, nil
	}

	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO questions
			(test_id, prompt, question_type, points, position, correct_text_answer)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, testID, item.Prompt, item.Type, item.Points, position, item.CorrectTextAnswer).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert question: %w", err)
	}
	return id, nil
}

//  двухпроходный: сначала все is_correct=false (обход частичного UNIQUE INDEX, который не DEFERRABLE), затем один -> true.
func upsertOptions(
	ctx context.Context,
	tx pgx.Tx,
	questionID uuid.UUID,
	opts []DraftOptionItem,
) error {
	keptIDs := make([]uuid.UUID, 0, len(opts))
	for _, opt := range opts {
		if opt.ID == nil {
			continue
		}
		parsed, err := uuid.Parse(*opt.ID)
		if err != nil {
			return fmt.Errorf("%w: invalid option id %q", ErrInvalidReference, *opt.ID)
		}
		keptIDs = append(keptIDs, parsed)
	}

	_, err := tx.Exec(ctx, `
		DELETE FROM answer_options
		WHERE question_id = $1
		  AND (cardinality($2::uuid[]) = 0 OR id <> ALL($2::uuid[]))
	`, questionID, keptIDs)
	if err != nil {
		return fmt.Errorf("delete removed options: %w", err)
	}

	// Первый проход: insert/update с is_correct=false
	var correctID *uuid.UUID
	for pos, opt := range opts {
		var optID uuid.UUID

		if opt.ID != nil {
			parsed, _ := uuid.Parse(*opt.ID)
			optID = parsed

			res, err := tx.Exec(ctx, `
				UPDATE answer_options
				SET option_text = $1, is_correct = false, position = $2
				WHERE id = $3 AND question_id = $4
			`, opt.Text, pos, optID, questionID)
			if err != nil {
				return fmt.Errorf("update option: %w", err)
			}
			if res.RowsAffected() == 0 {
				return fmt.Errorf(
					"%w: option %s does not belong to question %s",
					ErrInvalidReference, optID, questionID,
				)
			}
		} else {
			err := tx.QueryRow(ctx, `
				INSERT INTO answer_options (question_id, option_text, is_correct, position)
				VALUES ($1, $2, false, $3)
				RETURNING id
			`, questionID, opt.Text, pos).Scan(&optID)
			if err != nil {
				return fmt.Errorf("insert option: %w", err)
			}
		}

		if opt.IsCorrect && correctID == nil {
			cp := optID
			correctID = &cp
		}
	}

	// Второй проход: отметить правильный
	if correctID != nil {
		_, err := tx.Exec(ctx,
			`UPDATE answer_options SET is_correct = true WHERE id = $1`,
			*correctID)
		if err != nil {
			return fmt.Errorf("set correct option: %w", err)
		}
	}
	return nil
}
