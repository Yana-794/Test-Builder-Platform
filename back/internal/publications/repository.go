package publications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// тест + вопросы + варианты
type DraftForPublish struct {
	Test struct {
		ID          uuid.UUID
		Title       string
		Description string
		Subject     *string
		Status      string
		Revision    int
	}
	Questions []QuestionRow
}

type QuestionRow struct {
	ID                uuid.UUID
	Prompt            string
	Type              string
	Points            float64
	Position          int
	CorrectTextAnswer *string
	Options           []OptionRow
}

type OptionRow struct {
	ID        uuid.UUID
	Text      string
	IsCorrect bool
	Position  int
}

func (r *Repository) LoadDraftForPublish(
	ctx context.Context,
	q querier,
	ownerID, testID uuid.UUID,
) (DraftForPublish, error) {
	return r.loadDraft(ctx, q, ownerID, testID, true)
}

func (r *Repository) LoadDraftForValidation(
	ctx context.Context,
	q querier,
	ownerID, testID uuid.UUID,
) (DraftForPublish, error) {
	return r.loadDraft(ctx, q, ownerID, testID, false)
}

func (r *Repository) loadDraft(
	ctx context.Context,
	q querier,
	ownerID, testID uuid.UUID,
	forUpdate bool,
) (DraftForPublish, error) {
	var draft DraftForPublish

	query := `SELECT id, title, description, subject, status, revision
	          FROM tests
	          WHERE id = $1 AND owner_id = $2`
	if forUpdate {
		query += " FOR UPDATE"
	}

	err := q.QueryRow(ctx, query, testID, ownerID).Scan(
		&draft.Test.ID,
		&draft.Test.Title,
		&draft.Test.Description,
		&draft.Test.Subject,
		&draft.Test.Status,
		&draft.Test.Revision,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return DraftForPublish{}, ErrTestNotFound
	}
	if err != nil {
		return DraftForPublish{}, fmt.Errorf("load test: %w", err)
	}

	qrows, err := q.Query(ctx, `
		SELECT id, prompt, question_type, points::float8, position, correct_text_answer
		FROM questions
		WHERE test_id = $1
		ORDER BY position
	`, testID)
	if err != nil {
		return DraftForPublish{}, fmt.Errorf("query questions: %w", err)
	}
	defer qrows.Close()

	// нужен, чтобы при загрузке опций вторым запросом быстро найти, к какому вопросу относится опция, и добавить по индексу
	index := make(map[uuid.UUID]int)

	for qrows.Next() {
		var q QuestionRow
		if err := qrows.Scan(
			&q.ID, &q.Prompt, &q.Type, &q.Points, &q.Position, &q.CorrectTextAnswer,
		); err != nil {
			return DraftForPublish{}, fmt.Errorf("scan question: %w", err)
		}
		q.Options = make([]OptionRow, 0)
		index[q.ID] = len(draft.Questions)
		draft.Questions = append(draft.Questions, q)
	}
	if err := qrows.Err(); err != nil {
		return DraftForPublish{}, fmt.Errorf("rows questions: %w", err)
	}
	if len(draft.Questions) == 0 {
		return draft, nil
	}

	ids := make([]uuid.UUID, 0, len(draft.Questions))
	for _, q := range draft.Questions {
		ids = append(ids, q.ID)
	}

	orows, err := q.Query(ctx, `
		SELECT id, question_id, option_text, is_correct, position
		FROM answer_options
		WHERE question_id = ANY($1)
		ORDER BY question_id, position
	`, ids)
	if err != nil {
		return DraftForPublish{}, fmt.Errorf("query options: %w", err)
	}
	defer orows.Close()

	for orows.Next() {
		var o OptionRow
		var qID uuid.UUID
		if err := orows.Scan(&o.ID, &qID, &o.Text, &o.IsCorrect, &o.Position); err != nil {
			return DraftForPublish{}, fmt.Errorf("scan option: %w", err)
		}
		if i, ok := index[qID]; ok {
			draft.Questions[i].Options = append(draft.Questions[i].Options, o)
		}
	}
	if err := orows.Err(); err != nil {
		return DraftForPublish{}, fmt.Errorf("rows options: %w", err)
	}

	return draft, nil
}

func (r *Repository) NextVersion(ctx context.Context, q querier, testID uuid.UUID) (int, error) {
	var version int
	err := q.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1
		FROM test_publications
		WHERE test_id = $1
	`, testID).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("query next version: %w", err)
	}
	return version, nil
}

func (r *Repository) CreatePublication(
	ctx context.Context,
	q querier,
	testID uuid.UUID,
	version int,
	publicToken string,
	content SnapshotContent,
) (Publication, error) {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return Publication{}, fmt.Errorf("marshal snapshot: %w", err)
	}

	var public Publication
	err = q.QueryRow(ctx, `
		INSERT INTO test_publications (test_id, version, public_token, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, test_id, version, public_token, created_at
	`, testID, version, publicToken, contentJSON).Scan(
		&public.ID, &public.TestID, &public.Version, &public.PublicToken, &public.CreatedAt,
	)
	if err != nil {
		return Publication{}, fmt.Errorf("insert publication: %w", err)
	}
	public.Content = content

	_, err = q.Exec(ctx, `
		UPDATE tests
		SET status = 'published', publication_version_id = $1, revision = revision + 1
		WHERE id = $2
	`, public.ID, testID)
	if err != nil {
		return Publication{}, fmt.Errorf("update test: %w", err)
	}

	return public, nil
}
func (r *Repository) GetPublicationByTest(
	ctx context.Context,
	ownerID, testID uuid.UUID,
) (Publication, error) {
	var pub Publication
	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.test_id, p.version, p.public_token, p.content, p.created_at
		FROM test_publications p
		JOIN tests t ON t.publication_version_id = p.id
		WHERE t.id = $1 AND t.owner_id = $2
	`, testID, ownerID).Scan(
		&pub.ID,
		&pub.TestID,
		&pub.Version,
		&pub.PublicToken,
		&raw,
		&pub.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Publication{}, ErrTestNotFound
	}
	if err != nil {
		return Publication{}, fmt.Errorf("get publication: %w", err)
	}
	_ = raw
	return pub, nil
}
