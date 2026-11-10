//go:build integration

package attempts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"test-builder/internal/attempts"
	"test-builder/internal/publications"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPool *pgxpool.Pool
	testRepo *attempts.Repository
	testSvc  *attempts.Service

	publicationID uuid.UUID
	publicToken   = "test-token-abc"

	qSingleID    uuid.UUID
	optCorrectID uuid.UUID
	optWrongID   uuid.UUID
	qTextID      uuid.UUID
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	initScript, cleanup, err := prepareMigrations()
	if err != nil {
		panic(fmt.Errorf("prepare migrations: %w", err))
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
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable",
		host, mappedPort.Port())

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Errorf("pgxpool: %w", err))
	}
	if err := testPool.Ping(ctx); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}
	testRepo = attempts.NewRepository(testPool)
	testSvc = attempts.NewService(testRepo, testPool)

	if err := seedPublication(ctx); err != nil {
		panic(fmt.Errorf("seed publication: %w", err))
	}

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
	tmp, err := os.CreateTemp("", "attempts-mig-*.sql")
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

func seedPublication(ctx context.Context) error {
	// user
	var userID uuid.UUID
	if err := testPool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, password_hash)
		VALUES ('attempts-test@example.com', 'Test', 'x')
		RETURNING id
	`).Scan(&userID); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	var testID uuid.UUID
	if err := testPool.QueryRow(ctx, `
		INSERT INTO tests (owner_id, title, description, status)
		VALUES ($1, 'Attempts Test', '', 'published')
		RETURNING id
	`, userID).Scan(&testID); err != nil {
		return fmt.Errorf("insert test: %w", err)
	}

	// snapshot
	qSingleID = uuid.New()
	optCorrectID = uuid.New()
	optWrongID = uuid.New()
	qTextID = uuid.New()

	correct := "Функция с лексическим окружением"

	snapshot := publications.SnapshotContent{
		SchemaVersion: 1,
		Title:         "Attempts Test",
		Description:   "",
		Questions: []publications.SnapshotQuestion{
			{
				ID:       qSingleID,
				Prompt:   "Сколько будет 1+1?",
				Type:     publications.TypeSingleChoice,
				Points:   1,
				Position: 0,
				Options: []publications.SnapshotOption{
					{ID: optCorrectID, Text: "2", IsCorrect: true, Position: 0},
					{ID: optWrongID, Text: "3", IsCorrect: false, Position: 1},
				},
			},
			{
				ID:                qTextID,
				Prompt:            "Что такое замыкание?",
				Type:              publications.TypeText,
				Points:            2,
				Position:          1,
				CorrectTextAnswer: &correct,
			},
		},
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	if err := testPool.QueryRow(ctx, `
		INSERT INTO test_publications (test_id, version, public_token, content)
		VALUES ($1, 1, $2, $3)
		RETURNING id
	`, testID, publicToken, raw).Scan(&publicationID); err != nil {
		return fmt.Errorf("insert publication: %w", err)
	}
	return nil
}

func cleanAttempts(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), `TRUNCATE attempts CASCADE`)
	require.NoError(t, err)
}

func dummyHash() []byte {
	return []byte{1, 2, 3, 4}
}

func newAttempt(t *testing.T) attempts.Attempt {
	t.Helper()
	a, err := testRepo.CreateAttempt(context.Background(),
		publicationID, "Анна", dummyHash(), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	return a
}

func TestRepository_GetPublicationByToken_Success(t *testing.T) {
	pws, err := testRepo.GetPublicationByToken(context.Background(), publicToken)
	require.NoError(t, err)

	assert.Equal(t, publicationID, pws.PublicationID)
	assert.Equal(t, "Attempts Test", pws.Content.Title)
	assert.Len(t, pws.Content.Questions, 2)
	assert.Equal(t, "single_choice", pws.Content.Questions[0].Type)
	assert.Equal(t, "text", pws.Content.Questions[1].Type)
}

func TestRepository_GetPublicationByToken_NotFound(t *testing.T) {
	_, err := testRepo.GetPublicationByToken(context.Background(), "missing-token")
	assert.ErrorIs(t, err, attempts.ErrPublicationNotFound)
}

func TestRepository_GetPublicationByID_Success(t *testing.T) {
	content, err := testRepo.GetPublicationByID(context.Background(), publicationID)
	require.NoError(t, err)
	assert.Equal(t, "Attempts Test", content.Title)
	assert.Len(t, content.Questions, 2)
}

func TestRepository_CreateAttempt(t *testing.T) {
	cleanAttempts(t)

	expires := time.Now().UTC().Add(24 * time.Hour)
	hash := []byte{9, 9, 9}
	a, err := testRepo.CreateAttempt(context.Background(),
		publicationID, "Анна", hash, expires)
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, a.ID)
	assert.Equal(t, publicationID, a.PublicationID)
	assert.Equal(t, "Анна", a.StudentName)
	assert.Equal(t, attempts.StatusInProgress, a.Status)
	assert.Equal(t, hash, a.AccessTokenHash)
	assert.WithinDuration(t, time.Now(), a.StartedAt, 5*time.Second)
	assert.WithinDuration(t, expires, a.ExpiresAt, time.Second)

	assert.Nil(t, a.CompletedAt)
	assert.Nil(t, a.Score)
	assert.Nil(t, a.MaxScore)
	assert.Nil(t, a.ScorePercent)
	assert.Nil(t, a.DurationSeconds)
}

func TestRepository_GetAttemptByID_Success(t *testing.T) {
	cleanAttempts(t)
	created := newAttempt(t)

	got, err := testRepo.GetAttemptByID(context.Background(), created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.StudentName, got.StudentName)
	assert.Equal(t, attempts.StatusInProgress, got.Status)
}

func TestRepository_GetAttemptByID_NotFound(t *testing.T) {
	cleanAttempts(t)
	_, err := testRepo.GetAttemptByID(context.Background(), uuid.New())
	assert.ErrorIs(t, err, attempts.ErrAttemptNotFound)
}

func TestRepository_SaveAnswers_Insert(t *testing.T) {
	cleanAttempts(t)
	a := newAttempt(t)

	txt := "yes"
	correctOptID := optCorrectID
	err := saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
		{QuestionID: qTextID, AnswerText: &txt},
	})
	require.NoError(t, err)

	got, err := testRepo.GetAnswers(context.Background(), testPool, a.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestRepository_SaveAnswers_Replace(t *testing.T) {
	cleanAttempts(t)
	a := newAttempt(t)

	correctOptID := optCorrectID
	txt := "yes"
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
		{QuestionID: qTextID, AnswerText: &txt},
	}))

	newTxt := "функция"
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qTextID, AnswerText: &newTxt},
	}))

	got, err := testRepo.GetAnswers(context.Background(), testPool, a.ID)
	require.NoError(t, err)
	require.Len(t, got, 1, "второй вопрос должен быть удалён")
	assert.Equal(t, qTextID, got[0].QuestionID)
	require.NotNil(t, got[0].AnswerText)
	assert.Equal(t, "функция", *got[0].AnswerText)
}

func TestRepository_SaveAnswers_ReplaceWithEmpty(t *testing.T) {
	cleanAttempts(t)
	a := newAttempt(t)

	correctOptID := optCorrectID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
	}))

	require.NoError(t, saveAnswers(t, a.ID, nil))

	got, err := testRepo.GetAnswers(context.Background(), testPool, a.ID)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func saveAnswers(t *testing.T, attemptID uuid.UUID, answers []attempts.Answer) error {
	t.Helper()
	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()
	if err := testRepo.SaveAnswers(ctx, tx, attemptID, answers); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func TestRepository_CompleteAttempt(t *testing.T) {
	cleanAttempts(t)
	a := newAttempt(t)

	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(ctx) }()

	completedAt := time.Now().UTC()
	err = testRepo.CompleteAttempt(ctx, tx, a.ID, 8, 10, 80, 150, completedAt)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))

	got, err := testRepo.GetAttemptByID(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, attempts.StatusCompleted, got.Status)
	require.NotNil(t, got.Score)
	assert.Equal(t, 8.0, *got.Score)
	require.NotNil(t, got.MaxScore)
	assert.Equal(t, 10.0, *got.MaxScore)
	require.NotNil(t, got.ScorePercent)
	assert.Equal(t, 80.0, *got.ScorePercent)
	require.NotNil(t, got.DurationSeconds)
	assert.Equal(t, 150, *got.DurationSeconds)
	require.NotNil(t, got.CompletedAt)
	assert.WithinDuration(t, completedAt, *got.CompletedAt, time.Second)
}

func TestService_Submit_AllCorrect(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	txt := "  функция с лексическим окружением  " // разные регистр/trim — должно засчитаться
	correctOptID := optCorrectID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
		{QuestionID: qTextID, AnswerText: &txt},
	}))

	result, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	assert.Equal(t, attempts.StatusCompleted, result.Status)
	assert.Equal(t, 3.0, result.Score, "1 (single) + 2 (text) = 3")
	assert.Equal(t, 3.0, result.MaxScore)
	assert.Equal(t, 100.0, result.Percentage)
}

func TestService_Submit_AllWrong(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Иван")
	require.NoError(t, err)

	wrongOptID := optWrongID
	txt := "неправильный ответ"
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &wrongOptID},
		{QuestionID: qTextID, AnswerText: &txt},
	}))

	result, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	assert.Equal(t, 0.0, result.Score)
	assert.Equal(t, 3.0, result.MaxScore)
	assert.Equal(t, 0.0, result.Percentage)
}

func TestService_Submit_PartialAnswers(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Мария")
	require.NoError(t, err)

	correctOptID := optCorrectID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
	}))

	result, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	assert.Equal(t, 1.0, result.Score, "только single_choice")
	assert.Equal(t, 3.0, result.MaxScore)
	assert.InDelta(t, 33.33, result.Percentage, 0.01)
}

func TestService_Submit_NoAnswers(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Пётр")
	require.NoError(t, err)

	result, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	assert.Equal(t, 0.0, result.Score)
	assert.Equal(t, 3.0, result.MaxScore)
	assert.Equal(t, 0.0, result.Percentage)
}

func TestService_Submit_Idempotent(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	correctOptID := optCorrectID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
	}))

	r1, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)
	assert.Equal(t, 1.0, r1.Score)

	wrongOptID := optWrongID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &wrongOptID},
	}))

	r2, err := testSvc.Submit(context.Background(), a)
	require.NoError(t, err)
	assert.Equal(t, r1.Score, r2.Score, "идемпотентность: score не должен меняться")
	assert.Equal(t, r1.Percentage, r2.Percentage)
	assert.Equal(t, r1.CompletedAt, r2.CompletedAt, "completed_at не должен обновляться")
}

func TestService_SaveAnswers_RejectForeignQuestionID(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	txt := "x"
	err = testSvc.SaveAnswers(context.Background(), a, attempts.SaveAnswersRequest{
		Answers: []attempts.AnswerDTO{
			{QuestionID: uuid.NewString(), AnswerText: &txt},
		},
	})
	require.Error(t, err)

	var ve *attempts.ValidationError
	require.ErrorAs(t, err, &ve)
	require.NotEmpty(t, ve.Details)
}

func TestService_SaveAnswers_RejectWrongOptionForQuestion(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	// optCorrectID не принадлежит чужому вопросу — используем не тот optionId
	// но передаём qTextID (text) + selectedOptionId (нельзя) → ValidationError
	wrong := optCorrectID.String()
	err = testSvc.SaveAnswers(context.Background(), a, attempts.SaveAnswersRequest{
		Answers: []attempts.AnswerDTO{
			{QuestionID: qTextID.String(), SelectedOptionID: &wrong},
		},
	})
	require.Error(t, err)

	var ve *attempts.ValidationError
	require.ErrorAs(t, err, &ve)
}

func TestService_SaveAnswers_RejectAfterCompleted(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	_, err = testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	// после submit — попытка completed, SaveAnswers → ErrAttemptCompleted
	txt := "x"
	err = testSvc.SaveAnswers(context.Background(), a, attempts.SaveAnswersRequest{
		Answers: []attempts.AnswerDTO{
			{QuestionID: qTextID.String(), AnswerText: &txt},
		},
	})
	assert.ErrorIs(t, err, attempts.ErrAttemptCompleted)
}

func TestService_GetResult_Completed(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	correctOptID := optCorrectID
	require.NoError(t, saveAnswers(t, a.ID, []attempts.Answer{
		{QuestionID: qSingleID, SelectedOptionID: &correctOptID},
	}))

	_, err = testSvc.Submit(context.Background(), a)
	require.NoError(t, err)

	result, err := testSvc.GetResult(context.Background(), a)
	require.NoError(t, err)
	assert.Equal(t, attempts.StatusCompleted, result.Status)
	assert.Equal(t, 1.0, result.Score)
}

func TestService_GetResult_NotCompleted(t *testing.T) {
	cleanAttempts(t)
	a, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "Анна")
	require.NoError(t, err)

	_, err = testSvc.GetResult(context.Background(), a)
	assert.ErrorIs(t, err, attempts.ErrAttemptCompleted)
}

func TestService_CreateAttempt_EmptyName(t *testing.T) {
	cleanAttempts(t)
	_, _, err := testSvc.CreateAttempt(context.Background(), publicToken, "   ")
	require.Error(t, err)

	var ve *attempts.ValidationError
	require.ErrorAs(t, err, &ve)
	require.NotEmpty(t, ve.Details)
	assert.Equal(t, "studentName", ve.Details[0].Field)
}

func TestService_CreateAttempt_NameTooLong(t *testing.T) {
	cleanAttempts(t)
	long := strings.Repeat("a", 121)
	_, _, err := testSvc.CreateAttempt(context.Background(), publicToken, long)
	require.Error(t, err)

	var ve *attempts.ValidationError
	require.ErrorAs(t, err, &ve)
}

func TestService_CreateAttempt_PublicationNotFound(t *testing.T) {
	cleanAttempts(t)
	_, _, err := testSvc.CreateAttempt(context.Background(), "missing-token", "Анна")
	assert.ErrorIs(t, err, attempts.ErrPublicationNotFound)
}
