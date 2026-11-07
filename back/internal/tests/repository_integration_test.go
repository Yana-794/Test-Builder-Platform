//go:build integration

package tests_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"test-builder/internal/tests"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPool *pgxpool.Pool
	testRepo *tests.Repository
	userA    uuid.UUID
	userB    uuid.UUID
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
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, mappedPort.Port())

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic(err)
	}
	if err := testPool.Ping(ctx); err != nil {
		panic(err)
	}
	testRepo = tests.NewRepository(testPool)

	userA = insertUser(ctx, "a@example.com")
	userB = insertUser(ctx, "b@example.com")

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

func cleanTests(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE tests CASCADE`)
	require.NoError(t, err)
}



func TestRepository_CreateTest(t *testing.T) {
	cleanTests(t)

	subject := "Programming"
	got, err := testRepo.CreateTest(context.Background(), userA, "JS Basics", "desc", &subject)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, got.ID)
	assert.Equal(t, userA, got.OwnerID)
	assert.Equal(t, "draft", string(got.Status))
	assert.Equal(t, 1, got.Revision)
	assert.Equal(t, "JS Basics", got.Title)
	require.NotNil(t, got.Subject)
	assert.Equal(t, "Programming", *got.Subject)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, 5*time.Second)
}



func TestRepository_GetTest_Isolation(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	got, qs, err := testRepo.GetTestWithQuestions(context.Background(), userA, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.NotNil(t, qs)

	_, _, err = testRepo.GetTestWithQuestions(context.Background(), userB, created.ID)
	assert.ErrorIs(t, err, tests.ErrTestNotFound)
}



func ptrStr(s string) *string { return &s }

func TestRepository_SaveDraft_InsertNew(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	req := tests.SaveDraftRequest{
		ExpectedRevision: 1,
		Test:             tests.DraftTest{Title: "T", Description: ""},
		Questions: []tests.DraftItem{
			{
				Prompt: "Q1?", Type: "single_choice", Points: 1,
				Options: []tests.DraftOptionItem{
					{Text: "a", IsCorrect: false},
					{Text: "b", IsCorrect: true},
				},
			},
			{
				Prompt: "Q2?", Type: "text", Points: 2.5,
				CorrectTextAnswer: ptrStr("answer"),
			},
		},
	}

	saved, qs, err := testRepo.SaveDraft(context.Background(), userA, created.ID, 1, req)
	require.NoError(t, err)

	assert.Equal(t, 2, saved.Revision)
	require.Len(t, qs, 2)

	
	assert.Equal(t, "single_choice", string(qs[0].Type))
	assert.Equal(t, 0, qs[0].Position)
	assert.Len(t, qs[0].Options, 2)
	assert.False(t, qs[0].Options[0].IsCorrect)
	assert.True(t, qs[0].Options[1].IsCorrect)
	assert.NotEqual(t, uuid.Nil, qs[0].Options[0].ID)

	assert.Equal(t, "text", string(qs[1].Type))
	assert.Equal(t, 1, qs[1].Position)
	require.NotNil(t, qs[1].CorrectTextAnswer)
	assert.Equal(t, "answer", *qs[1].CorrectTextAnswer)
	assert.Equal(t, 2.5, qs[1].Points)
}

func TestRepository_SaveDraft_UpdateExistingAndDeleteRemoved(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	_, qs1, err := testRepo.SaveDraft(context.Background(), userA, created.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1,
		Test:             tests.DraftTest{Title: "T"},
		Questions: []tests.DraftItem{
			{Prompt: "Q1", Type: "single_choice", Points: 1, Options: []tests.DraftOptionItem{
				{Text: "a", IsCorrect: false}, {Text: "b", IsCorrect: true},
			}},
			{Prompt: "Q2", Type: "text", Points: 1, CorrectTextAnswer: ptrStr("x")},
			{Prompt: "Q3", Type: "text", Points: 1, CorrectTextAnswer: ptrStr("y")},
		},
	})
	require.NoError(t, err)
	require.Len(t, qs1, 3)
	q1ID := qs1[0].ID
	q1Opt0ID := qs1[0].Options[0].ID
	q1Opt1ID := qs1[0].Options[1].ID

	saved, qs2, err := testRepo.SaveDraft(context.Background(), userA, created.ID, 2, tests.SaveDraftRequest{
		ExpectedRevision: 2,
		Test:             tests.DraftTest{Title: "T updated"},
		Questions: []tests.DraftItem{
			{
				ID: ptrStr(q1ID.String()), Prompt: "Q1 updated", Type: "single_choice", Points: 1.5,
				Options: []tests.DraftOptionItem{
					{ID: ptrStr(q1Opt1ID.String()), Text: "b", IsCorrect: true},
					{Text: "c", IsCorrect: false},
				},
			},
			{
				ID: ptrStr(qs1[2].ID.String()), Prompt: "Q3", Type: "text", Points: 1,
				CorrectTextAnswer: ptrStr("y"),
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 3, saved.Revision)
	require.Len(t, qs2, 2)
	assert.Equal(t, q1ID, qs2[0].ID)
	assert.Equal(t, "Q1 updated", qs2[0].Prompt)
	assert.Len(t, qs2[0].Options, 2)
	
	for _, o := range qs2[0].Options {
		assert.NotEqual(t, q1Opt0ID, o.ID)
	}
	var cnt int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM questions WHERE test_id = $1`, created.ID).Scan(&cnt))
	assert.Equal(t, 2, cnt)
}

func TestRepository_SaveDraft_RevisionConflict(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	_, _, err = testRepo.SaveDraft(context.Background(), userA, created.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "T"},
	})
	require.NoError(t, err)

	_, _, err = testRepo.SaveDraft(context.Background(), userA, created.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "T2"},
	})
	assert.ErrorIs(t, err, tests.ErrRevisionConflict)
}

func TestRepository_SaveDraft_NotDraft(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	_, err = testPool.Exec(context.Background(),
		`UPDATE tests SET status = 'published' WHERE id = $1`, created.ID)
	require.NoError(t, err)

	_, _, err = testRepo.SaveDraft(context.Background(), userA, created.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "T"},
	})
	assert.ErrorIs(t, err, tests.ErrTestNotDraft)
}

func TestRepository_SaveDraft_InvalidReference(t *testing.T) {
	cleanTests(t)

	tA, err := testRepo.CreateTest(context.Background(), userA, "A", "", nil)
	require.NoError(t, err)
	_, qsA, err := testRepo.SaveDraft(context.Background(), userA, tA.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "A"},
		Questions: []tests.DraftItem{
			{Prompt: "q", Type: "text", Points: 1, CorrectTextAnswer: ptrStr("x")},
		},
	})
	require.NoError(t, err)
	qForeign := qsA[0].ID

	tB, err := testRepo.CreateTest(context.Background(), userA, "B", "", nil)
	require.NoError(t, err)

	_, _, err = testRepo.SaveDraft(context.Background(), userA, tB.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "B"},
		Questions: []tests.DraftItem{
			{ID: ptrStr(qForeign.String()), Prompt: "stolen", Type: "text", Points: 1,
				CorrectTextAnswer: ptrStr("x")},
		},
	})
	assert.ErrorIs(t, err, tests.ErrInvalidReference)

	var cnt int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM questions WHERE test_id = $1`, tB.ID).Scan(&cnt))
	assert.Equal(t, 0, cnt)
}

func TestRepository_SaveDraft_NotFoundForOtherUser(t *testing.T) {
	cleanTests(t)

	created, err := testRepo.CreateTest(context.Background(), userA, "T", "", nil)
	require.NoError(t, err)

	_, _, err = testRepo.SaveDraft(context.Background(), userB, created.ID, 1, tests.SaveDraftRequest{
		ExpectedRevision: 1, Test: tests.DraftTest{Title: "hacked"},
	})
	assert.ErrorIs(t, err, tests.ErrTestNotFound)
}