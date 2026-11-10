package attempts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"test-builder/internal/publications"

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

type PublicationWithSnapshot struct {
	PublicationID uuid.UUID
	Content       publications.SnapshotContent
}

func (r *Repository) GetPublicationByToken(ctx context.Context, token string) (PublicationWithSnapshot, error) {
	var pws PublicationWithSnapshot
	var raw []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, content 
		FROM test_publications 
		WHERE public_token = $1
		`, token).Scan(&pws.PublicationID, &raw)

	if errors.Is(err, pgx.ErrNoRows) {
		return PublicationWithSnapshot{}, ErrPublicationNotFound
	}
	if err != nil {
		return PublicationWithSnapshot{}, fmt.Errorf("get publication: %w", err)
	}
	if err := json.Unmarshal(raw, &pws.Content); err != nil {
		return PublicationWithSnapshot{}, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return pws, nil
}

func (r *Repository) GetPublicationByID(ctx context.Context, id uuid.UUID) (publications.SnapshotContent, error) {
	var raw []byte

	err := r.pool.QueryRow(ctx, `SELECT content FROM test_publications WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		return publications.SnapshotContent{}, fmt.Errorf("get publication: %w", err)
	}
	var c publications.SnapshotContent
	if err := json.Unmarshal(raw, &c); err != nil {
		return publications.SnapshotContent{}, fmt.Errorf("unmarshal: %w", err)
	}
	return c, nil
}

func (r *Repository) CreateAttempt(
	ctx context.Context,
	publicationID uuid.UUID,
	studentName string,
	accessTokenHash []byte,
	expiresAt time.Time,
) (Attempt, error) {
	var a Attempt
	err := r.pool.QueryRow(ctx, `
		INSERT INTO attempts (
		publication_id, 
		student_name, 
		status, 
		access_token_hash,
		started_at, expires_at)
		VALUES ($1, $2, 'in_progress', $3, now(), $4)
		RETURNING 
		id, 
		publication_id, 
		student_name, 
		status, 
		access_token_hash, 
		started_at, 
		expires_at
	`, publicationID,
		studentName,
		accessTokenHash,
		expiresAt).Scan(
		&a.ID,
		&a.PublicationID,
		&a.StudentName,
		&a.Status,
		&a.AccessTokenHash,
		&a.StartedAt,
		&a.ExpiresAt,
	)
	if err != nil {
		return Attempt{}, fmt.Errorf("create attempt: %w", err)
	}
	return a, nil
}

func (r *Repository) getAttempt(
	ctx context.Context,
	q querier,
	id uuid.UUID,
	forUpdate bool,
) (Attempt, error) {
	query := `
		SELECT  id, 
				publication_id, 
				student_name, status, 
				access_token_hash,
		       started_at,
			    expires_at, 
				completed_at,
		       score::float8,
			    max_score::float8, 
				score_percent::float8,
				 duration_seconds
		FROM attempts WHERE id = $1`
	if forUpdate {
		query += " FOR UPDATE"
	}

	var a Attempt
	err := q.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.PublicationID,
		&a.StudentName,
		&a.Status,
		&a.AccessTokenHash,
		&a.StartedAt,
		&a.ExpiresAt,
		&a.CompletedAt,
		&a.Score,
		&a.MaxScore,
		&a.ScorePercent,
		&a.DurationSeconds,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Attempt{}, ErrAttemptNotFound
	}
	if err != nil {
		return Attempt{}, fmt.Errorf("get attempt: %w", err)
	}
	return a, nil
}

func (r *Repository) GetAttemptByID(ctx context.Context, id uuid.UUID) (Attempt, error) {
	return r.getAttempt(ctx, r.pool, id, false)
}

func (r *Repository) LockAttempt(ctx context.Context, q querier, id uuid.UUID) (Attempt, error) {
	return r.getAttempt(ctx, q, id, true)
}

func (r *Repository) GetAnswers(ctx context.Context, q querier, attemptID uuid.UUID) ([]Answer, error) {
	rows, err := q.Query(ctx, `
		SELECT question_id, answer_text, selected_option_id
		FROM attempt_answers 
		WHERE attempt_id = $1
	`, attemptID)
	if err != nil {
		return nil, fmt.Errorf("query answers: %w", err)
	}
	defer rows.Close()

	var out []Answer
	for rows.Next() {
		var a Answer
		if err := rows.Scan(&a.QuestionID, &a.AnswerText, &a.SelectedOptionID); err != nil {
			return nil, fmt.Errorf("scan answer: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) SaveAnswers(ctx context.Context, tx pgx.Tx, attemptID uuid.UUID, answers []Answer) error {
	kept := make([]uuid.UUID, 0, len(answers))
	for _, a := range answers {
		kept = append(kept, a.QuestionID)
	}
	_, err := tx.Exec(ctx, `
		DELETE FROM attempt_answers
		WHERE attempt_id = $1
		  AND (cardinality($2::uuid[]) = 0 OR question_id <> ALL($2::uuid[]))
	`, attemptID, kept)
	if err != nil {
		return fmt.Errorf("delete removed answers: %w", err)
	}

	for _, a := range answers {
		_, err := tx.Exec(ctx, `
			INSERT INTO attempt_answers (attempt_id, question_id, answer_text, selected_option_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (attempt_id, question_id)
			DO UPDATE SET answer_text = EXCLUDED.answer_text,
			              selected_option_id = EXCLUDED.selected_option_id
		`, attemptID, a.QuestionID, a.AnswerText, a.SelectedOptionID)
		if err != nil {
			return fmt.Errorf("upsert answer: %w", err)
		}
	}
	return nil
}

func (r *Repository) CompleteAttempt(
	ctx context.Context, tx pgx.Tx, id uuid.UUID,
	score, maxScore, percent float64, duration int, completedAt time.Time,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE attempts
		SET status = 'completed', score = $1, max_score = $2, score_percent = $3,
		    duration_seconds = $4, completed_at = $5
		WHERE id = $6
	`, score, maxScore, percent, duration, completedAt, id)
	if err != nil {
		return fmt.Errorf("complete attempt: %w", err)
	}
	return nil
}
