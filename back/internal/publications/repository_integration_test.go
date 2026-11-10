//go:build integration

package publications_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"encoding/json" 
	"test-builder/internal/publications"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPool *pgxpool.Pool
	testSvc  *publications.Service
	userA    uuid.UUID
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	initScript, cleanup, err := prepareMigrations()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(initScript),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic(err)
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()

	host, err := pgContainer.Host(ctx)
	if err != nil {
		panic(err)
	}
	mappedPort, err := pgContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(err)
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
		panic(err)
	}
	if err := testPool.Ping(ctx); err != nil {
		panic(err)
	}

	repo := publications.NewRepository(testPool)
	testSvc = publications.NewService(repo, testPool, "https://example.com")

	userA = insertUser(ctx, "a@example.com")

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func prepareMigrations() (string, func(), error) {
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, err
	}
	var b strings.Builder
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return "", nil, err
		}
		up := strings.SplitN(string(data), "-- +goose Down", 2)[0]
		up = strings.ReplaceAll(up, "-- +goose Up", "")
		up = strings.ReplaceAll(up, "-- +goose StatementBegin", "")
		up = strings.ReplaceAll(up, "-- +goose StatementEnd", "")
		b.WriteString(up)
		b.WriteString("\n")
	}
	tmp, err := os.CreateTemp("", "mig-*.sql")
	if err != nil {
		return "", nil, err
	}
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		return "", nil, err
	}
	tmp.Close()
	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}

func insertUser(ctx context.Context, email string) uuid.UUID {
	var id uuid.UUID
	err := testPool.QueryRow(ctx,
		`INSERT INTO users (email, display_name, password_hash) VALUES ($1,$2,$3) RETURNING id`,
		email, "u-"+email, "x",
	).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func cleanDB(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE tests CASCADE`)
	require.NoError(t, err)
}

func seedPublishableTest(t *testing.T, ownerID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	var testID uuid.UUID
	err := testPool.QueryRow(ctx, `
		INSERT INTO tests (owner_id, title, description, status, revision)
		VALUES ($1, 'Test Title', 'Test Description', 'draft', 1)
		RETURNING id
	`, ownerID).Scan(&testID)
	require.NoError(t, err)

	_, err = testPool.Exec(ctx, `
		INSERT INTO questions (test_id, prompt, question_type, points, position, correct_text_answer)
		VALUES ($1, 'What is Go?', 'text', 5, 0, 'A language')
	`, testID)
	require.NoError(t, err)

	return testID
}

func TestService_Publish_Success(t *testing.T) {
	cleanDB(t)
	testID := seedPublishableTest(t, userA)

	pub, err := testSvc.Publish(context.Background(), userA, testID, 1)
	require.NoError(t, err)

	assert.NotEmpty(t, pub.Publication.ID, "id публикации должен быть сгенерирован")
	assert.Equal(t, 1, pub.Publication.Version, "первая публикация — version 1")
	assert.NotEmpty(t, pub.Publication.PublicToken, "токен должен быть сгенерирован")
	assert.Contains(t, pub.Publication.PublicURL, "/test/", "URL должен содержать /test/{token}")
	assert.NotEmpty(t, pub.Publication.PublishedAt, "дата публикации должна быть заполнена")

	var status string
	var revision int
	err = testPool.QueryRow(context.Background(), `
		SELECT status, revision FROM tests WHERE id = $1
	`, testID).Scan(&status, &revision)
	require.NoError(t, err)
	assert.Equal(t, "published", status, "статус должен стать published")
	assert.Equal(t, 2, revision, "revision должна инкрементироваться")

	var contentJSON string
	err = testPool.QueryRow(context.Background(), `
    SELECT content::text FROM test_publications WHERE test_id = $1
`, testID).Scan(&contentJSON)
	require.NoError(t, err)

	var snapshot struct {
		Title     string `json:"title"`
		Questions []struct {
			Prompt            string  `json:"prompt"`
			Points            float64 `json:"points"`
			CorrectTextAnswer *string `json:"correctTextAnswer"`
		} `json:"questions"`
	}
	require.NoError(t, json.Unmarshal([]byte(contentJSON), &snapshot), "content должен быть валидным JSON")

	assert.Equal(t, "Test Title", snapshot.Title, "снапшот должен содержать title")
	require.Len(t, snapshot.Questions, 1, "в снапшоте должен быть один вопрос")
	assert.Equal(t, "What is Go?", snapshot.Questions[0].Prompt, "вопрос сохранён в снапшоте")
	assert.Equal(t, 5.0, snapshot.Questions[0].Points)
	require.NotNil(t, snapshot.Questions[0].CorrectTextAnswer)
	assert.Equal(t, "A language", *snapshot.Questions[0].CorrectTextAnswer)
}
