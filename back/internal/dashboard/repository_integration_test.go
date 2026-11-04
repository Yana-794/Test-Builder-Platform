//go:build integration

package dashboard_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"test-builder/internal/dashboard"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const recentLimit = 5

var (
	testPool *pgxpool.Pool
	testRepo *dashboard.Repository
	userA    uuid.UUID
	userB    uuid.UUID
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	initScript, cleanup, err := prepareMigrations()
	if err != nil {
		panic(fmt.Errorf("prepare migrations: %w", err))
	}
	defer cleanup()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(initScript),
		postgres.BasicWaitStrategies(), 
	)
	if err != nil {
		panic(fmt.Errorf("start postgres: %w", err))
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()

	host, err := pgContainer.Host(ctx)
	if err != nil {
		panic(fmt.Errorf("container host: %w", err))
	}
	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(fmt.Errorf("mapped port: %w", err))
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}

	connStr := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/testdb?sslmode=disable",
		host, mappedPort.Port(),
	)

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Errorf("pgxpool: %w", err))
	}

	if err := testPool.Ping(ctx); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}

	testRepo = dashboard.NewRepository(testPool)

	userA = insertUser(ctx, "a@example.com")
	userB = insertUser(ctx, "b@example.com")

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}

func prepareMigrations() (string, func(), error) {
	// из back/internal/dashboard/ → back/migrations/
	migrationsDir := filepath.Join("..", "..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var combined strings.Builder
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(migrationsDir, e.Name()))
		if err != nil {
			return "", nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}

		up := strings.SplitN(string(data), "-- +goose Down", 2)[0]
		up = strings.ReplaceAll(up, "-- +goose Up", "")
		up = strings.ReplaceAll(up, "-- +goose StatementBegin", "")
		up = strings.ReplaceAll(up, "-- +goose StatementEnd", "")

		combined.WriteString(up)
		combined.WriteString("\n")
	}

	tmp, err := os.CreateTemp("", "migrations-*.sql")
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	if _, err := tmp.WriteString(combined.String()); err != nil {
		tmp.Close()
		return "", nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}

func insertUser(ctx context.Context, email string) uuid.UUID {
	var id uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO users (email, display_name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		email, "test-"+email, "not-a-real-hash",
	).Scan(&id)
	if err != nil {
		panic(fmt.Errorf("insert user %s: %w", email, err))
	}
	return id
}

func truncateTests(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE tests CASCADE`)
	require.NoError(t, err)
}

func insertTest(t *testing.T, owner uuid.UUID, title, status string, createdAt time.Time) {
	t.Helper()
	_, err := testPool.Exec(context.Background(),
		`INSERT INTO tests (owner_id, title, status, created_at) VALUES ($1, $2, $3, $4)`,
		owner, title, status, createdAt,
	)
	require.NoError(t, err)
}



func TestRepository_Empty(t *testing.T) {
	truncateTests(t)

	stats, err := testRepo.GetStats(context.Background(), userA)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalTests)
	assert.Equal(t, 0, stats.DraftTests)
	assert.Equal(t, 0, stats.PublishedTests)

	recent, err := testRepo.GetRecentTests(context.Background(), userA, recentLimit)
	require.NoError(t, err)
	assert.NotNil(t, recent, "recentTests должен быть [] а не nil")
	assert.Empty(t, recent)
}

func TestRepository_DraftsAndPublished(t *testing.T) {
	truncateTests(t)

	now := time.Now()
	insertTest(t, userA, "Draft 1", "draft", now.Add(-2*time.Hour))
	insertTest(t, userA, "Draft 2", "draft", now.Add(-1*time.Hour))
	insertTest(t, userA, "Published 1", "published", now)

	stats, err := testRepo.GetStats(context.Background(), userA)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalTests)
	assert.Equal(t, 2, stats.DraftTests)
	assert.Equal(t, 1, stats.PublishedTests)
}

func TestRepository_ExcludesArchived(t *testing.T) {
	truncateTests(t)

	insertTest(t, userA, "Active", "published", time.Now())
	insertTest(t, userA, "Archived", "archived", time.Now().Add(-time.Hour))

	stats, err := testRepo.GetStats(context.Background(), userA)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalTests)

	recent, err := testRepo.GetRecentTests(context.Background(), userA, recentLimit)
	require.NoError(t, err)
	require.Len(t, recent, 1)
	assert.Equal(t, "Active", recent[0].Title)
}

func TestRepository_RecentLimitAndSort(t *testing.T) {
	truncateTests(t)

	base := time.Now()
	for i := 1; i <= 7; i++ {
		insertTest(t, userA, fmt.Sprintf("Test %d", i), "draft",
			base.Add(time.Duration(i)*time.Hour))
	}

	recent, err := testRepo.GetRecentTests(context.Background(), userA, recentLimit)
	require.NoError(t, err)
	require.Len(t, recent, 5, "должен вернуться лимит 5")

	for i := 1; i < len(recent); i++ {
		assert.False(t, recent[i].CreatedAt.After(recent[i-1].CreatedAt),
			"порядок должен быть created_at DESC")
	}
}

func TestRepository_Isolation(t *testing.T) {
	truncateTests(t)

	insertTest(t, userA, "A test", "draft", time.Now())
	insertTest(t, userB, "B test", "published", time.Now())

	statsA, err := testRepo.GetStats(context.Background(), userA)
	require.NoError(t, err)
	assert.Equal(t, 1, statsA.TotalTests)
	assert.Equal(t, 1, statsA.DraftTests)

	statsB, err := testRepo.GetStats(context.Background(), userB)
	require.NoError(t, err)
	assert.Equal(t, 1, statsB.TotalTests)
	assert.Equal(t, 1, statsB.PublishedTests)
}