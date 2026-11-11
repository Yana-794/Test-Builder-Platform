package analytics

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositoty struct {
	pool *pgxpool.Pool
}

func NewRepository ( pool *pgxpool.Pool) *Repositoty {
	return  &Repositoty{pool: pool}
}
//  последние попытки по всем тестам текущего преподавателя
func (r *Repositoty) Recent( ctx context.Context, ownerID uuid.UUID, limit int) ([]RecentAttempt, error) {
	const q = `
	SELECT
			a.id,
			a.student_name,
			t.id,
			t.title,
			a.status,
			a.score::float8,
			a.max_score::float8,
			a.score_percent::float8,
			a.duration_seconds,
			a.started_at,
			a.completed_at
		FROM attempts a
		JOIN test_publications p ON p.id = a.publication_id
		JOIN tests t ON t.id = p.test_id
		WHERE t.owner_id = $1
		ORDER BY a.started_at DESC
		LIMIT $2
	` 
	rows, err := r.pool.Query(ctx, q, ownerID, limit) 
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]RecentAttempt, 0, limit)
	for rows.Next() {
		var a RecentAttempt
		if err := rows.Scan(
			&a.AttemptID, 
			&a.StudentName, 
			&a.TestID, 
			&a.TestTitle, 
			&a.Status,
			&a.Score, 
			&a.MaxScore, 
			&a.Percent,
			&a.DurationSeconds,
			&a.StartedAt, 
			&a.CompletedAt); err != nil {
				return  nil, fmt.Errorf("scan attempt: %w", err)
			}
			out = append(out, a)
	}
	return  out ,rows.Err()
}