package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetStats(ctx context.Context, ownerID uuid.UUID) (DashboardStats, error) {
	const q = `
	SELECT
    	COUNT(*) FILTER (WHERE status IN ('draft','published')) AS total_tests,
    	COUNT(*) FILTER (WHERE status = 'draft')                AS draft_tests,
    	COUNT(*) FILTER (WHERE status = 'published')            AS published_tests
	FROM tests
	WHERE owner_id = $1;
`
	var s DashboardStats
	err := r.pool.QueryRow(ctx, q, ownerID).Scan(
		&s.TotalTests, &s.DraftTests, &s.PublishedTests,
	)
	if err != nil {
		return DashboardStats{}, err
	}
	return s, nil
}

func (r *Repository) GetRecentTests(ctx context.Context, ownerID uuid.UUID, limit int) ([]RecentTest, error) {
	const q = `
	SELECT id, title, status, created_at
	FROM tests
	WHERE owner_id = $1
  		AND status IN ('draft','published')
	ORDER BY created_at DESC, id DESC
	LIMIT $2;
	`
	rows, err := r.pool.Query(ctx, q, ownerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RecentTest, 0, limit)
	for rows.Next() {
		var t RecentTest
		if err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
