package attempts

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"test-builder/internal/publications"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

// GET /public/tests/:token

// тест без правильных ответов
func (s *Service) GetPublicTest(ctx context.Context, token string) (publications.PublicTest, error) {
	pws, err := s.repo.GetPublicationByToken(ctx, token)
	if err != nil {
		return publications.PublicTest{}, err
	}
	return publications.ToPublicTest(pws.Content), nil
}

// POST /public/tests/:token/attempts

// создаёт попытку, возвращает её и секрет для cookie
func (s *Service) CreateAttempt(ctx context.Context, token, studentName string) (Attempt, string, error) {
	name := strings.TrimSpace(studentName)
	if name == "" {
		return Attempt{}, "", &ValidationError{Details: []FieldDetail{
			{Field: "studentName", Code: "REQUIRED", Message: "Имя обязательно"},
		}}
	}
	if utf8.RuneCountInString(name) > 120 {
		return Attempt{}, "", &ValidationError{Details: []FieldDetail{
			{Field: "studentName", Code: "TOO_LONG", Message: "Не более 120 символов"},
		}}
	}

	pws, err := s.repo.GetPublicationByToken(ctx, token)
	if err != nil {
		return Attempt{}, "", err
	}

	secretRaw := make([]byte, 32)
	if _, err := rand.Read(secretRaw); err != nil {
		return Attempt{}, "", fmt.Errorf("rand: %w", err)
	}
	secret := base64.RawURLEncoding.EncodeToString(secretRaw)

	expiresAt := time.Now().UTC().Add(AttemptLifetime)
	a, err := s.repo.CreateAttempt(ctx, pws.PublicationID, name, hashSecret(secret), expiresAt)
	if err != nil {
		return Attempt{}, "", err
	}
	return a, secret, nil
}

// аутентификация по cookie
// сравнивает секрет из cookie с хешем в бд
func (s *Service) AuthenticateAttempt(ctx context.Context, attemptID uuid.UUID, secret string) (Attempt, error) {
	a, err := s.repo.GetAttemptByID(ctx, attemptID)
	if err != nil {
		return Attempt{}, err
	}
	if subtle.ConstantTimeCompare(hashSecret(secret), a.AccessTokenHash) != 1 {
		return Attempt{}, ErrAttemptNotFound
	}
	return a, nil
}

// GET /public/attempts/:attemptId

func (s *Service) GetState(ctx context.Context, attempt Attempt) (AttemptStateResponse, error) {
	snapshot, err := s.repo.GetPublicationByID(ctx, attempt.PublicationID)
	if err != nil {
		return AttemptStateResponse{}, err
	}
	answers, err := s.repo.GetAnswers(ctx, s.pool, attempt.ID)
	if err != nil {
		return AttemptStateResponse{}, err
	}

	saved := make([]AnswerDTO, 0, len(answers))
	for _, a := range answers {
		var optID *string
		if a.SelectedOptionID != nil {
			s := a.SelectedOptionID.String()
			optID = &s
		}
		saved = append(saved, AnswerDTO{
			QuestionID:       a.QuestionID.String(),
			SelectedOptionID: optID,
			AnswerText:       a.AnswerText,
		})
	}

	return AttemptStateResponse{
		Attempt: AttemptShortDTO{
			ID:        attempt.ID.String(),
			Status:    attempt.Status,
			StartedAt: attempt.StartedAt.UTC().Format(time.RFC3339),
		},
		Test:    publications.ToPublicTest(snapshot),
		Answers: saved,
	}, nil
}

// PUT /public/attempts/:attemptId/answers

func (s *Service) SaveAnswers(ctx context.Context, attempt Attempt, req SaveAnswersRequest) error {
	if attempt.Status != StatusInProgress {
		return ErrAttemptCompleted
	}
	if time.Now().UTC().After(attempt.ExpiresAt) {
		return ErrAttemptExpired
	}

	snapshot, err := s.repo.GetPublicationByID(ctx, attempt.PublicationID)
	if err != nil {
		return err
	}

	parsed, verr := validateAnswers(snapshot, req.Answers)
	if verr != nil {
		return verr
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	locked, err := s.repo.LockAttempt(ctx, tx, attempt.ID)
	if err != nil {
		return err
	}
	if locked.Status != StatusInProgress {
		return ErrAttemptCompleted
	}

	if err := s.repo.SaveAnswers(ctx, tx, attempt.ID, parsed); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func validateAnswers(snapshot publications.SnapshotContent, inputs []AnswerDTO) ([]Answer, error) {
	qByID := make(map[uuid.UUID]publications.SnapshotQuestion, len(snapshot.Questions))
	for _, q := range snapshot.Questions {
		qByID[q.ID] = q
	}

	seen := make(map[uuid.UUID]bool, len(inputs))
	out := make([]Answer, 0, len(inputs))
	var details []FieldDetail

	for i, in := range inputs {
		field := fmt.Sprintf("answers[%d]", i)

		qid, err := uuid.Parse(in.QuestionID)
		if err != nil {
			details = append(details, FieldDetail{Field: field + ".questionId", Code: "INVALID", Message: "Невалидный UUID"})
			continue
		}
		q, ok := qByID[qid]
		if !ok {
			details = append(details, FieldDetail{Field: field + ".questionId", Code: "NOT_FOUND", Message: "Вопрос не из этого теста"})
			continue
		}
		if seen[qid] {
			details = append(details, FieldDetail{Field: field + ".questionId", Code: "DUPLICATE", Message: "Дублирующийся вопрос"})
			continue
		}
		seen[qid] = true

		a := Answer{QuestionID: qid}

		switch q.Type {
		case publications.TypeSingleChoice:
			if in.AnswerText != nil {
				details = append(details, FieldDetail{Field: field + ".answerText", Code: "INVALID", Message: "Не применимо для single_choice"})
				continue
			}
			if in.SelectedOptionID == nil {
				out = append(out, a) 
				continue
			}
			optID, err := uuid.Parse(*in.SelectedOptionID)
			if err != nil {
				details = append(details, FieldDetail{Field: field + ".selectedOptionId", Code: "INVALID", Message: "Невалидный UUID"})
				continue
			}
			found := false
			for _, o := range q.Options {
				if o.ID == optID {
					found = true
					break
				}
			}
			if !found {
				details = append(details, FieldDetail{Field: field + ".selectedOptionId", Code: "NOT_FOUND", Message: "Вариант не из этого вопроса"})
				continue
			}
			a.SelectedOptionID = &optID

		case publications.TypeText:
			if in.SelectedOptionID != nil {
				details = append(details, FieldDetail{Field: field + ".selectedOptionId", Code: "INVALID", Message: "Не применимо для text"})
				continue
			}
			if in.AnswerText != nil {
				a.AnswerText = in.AnswerText
			}
		}
		out = append(out, a)
	}

	if len(details) > 0 {
		return nil, &ValidationError{Details: details}
	}
	return out, nil
}

// POST /public/attempts/:attemptId/submit

func (s *Service) Submit(ctx context.Context, attempt Attempt) (AttemptResultDTO, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AttemptResultDTO{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	locked, err := s.repo.LockAttempt(ctx, tx, attempt.ID)
	if err != nil {
		return AttemptResultDTO{}, err
	}

	if locked.Status == StatusCompleted {
		return locked.ToResultDTO(), nil
	}

	if time.Now().UTC().After(locked.ExpiresAt) {
		return AttemptResultDTO{}, ErrAttemptExpired
	}

	snapshot, err := s.repo.GetPublicationByID(ctx, locked.PublicationID)
	if err != nil {
		return AttemptResultDTO{}, err
	}
	answers, err := s.repo.GetAnswers(ctx, tx, locked.ID)
	if err != nil {
		return AttemptResultDTO{}, err
	}

	score, maxScore := computeScore(snapshot, answers)
	var percent float64
	if maxScore > 0 {
		percent = math.Round(score/maxScore*10000) / 100
	}

	completedAt := time.Now().UTC()
	duration := int(completedAt.Sub(locked.StartedAt).Seconds())

	if err := s.repo.CompleteAttempt(ctx, tx, locked.ID, score, maxScore, percent, duration, completedAt); err != nil {
		return AttemptResultDTO{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return AttemptResultDTO{}, err
	}

	return AttemptResultDTO{
		ID:              locked.ID.String(),
		Status:          StatusCompleted,
		Score:           score,
		MaxScore:        maxScore,
		Percentage:      percent,
		DurationSeconds: duration,
		CompletedAt:     completedAt.Format(time.RFC3339),
	}, nil
}

// считает баллы по снапшоту (единственный источник правды)
func computeScore(snapshot publications.SnapshotContent, answers []Answer) (score, maxScore float64) {
	byQ := make(map[uuid.UUID]Answer, len(answers))
	for _, a := range answers {
		byQ[a.QuestionID] = a
	}

	for _, q := range snapshot.Questions {
		maxScore += q.Points

		a, ok := byQ[q.ID]
		if !ok {
			continue
		}

		switch q.Type {
		case publications.TypeSingleChoice:
			if a.SelectedOptionID == nil {
				continue
			}
			var correctID uuid.UUID
			for _, o := range q.Options {
				if o.IsCorrect {
					correctID = o.ID
					break
				}
			}
			if correctID == *a.SelectedOptionID {
				score += q.Points
			}

		case publications.TypeText:
			if a.AnswerText == nil || q.CorrectTextAnswer == nil {
				continue
			}
			if strings.EqualFold(
				strings.TrimSpace(*a.AnswerText),
				strings.TrimSpace(*q.CorrectTextAnswer),
			) {
				score += q.Points
			}
		}
	}
	return score, maxScore
}

// GET /public/attempts/:attemptId/result
func (s *Service) GetResult(ctx context.Context, attempt Attempt) (AttemptResultDTO, error) {
	a, err := s.repo.GetAttemptByID(ctx, attempt.ID)
	if err != nil {
		return AttemptResultDTO{}, err
	}
	if a.Status != StatusCompleted {
		return AttemptResultDTO{}, ErrAttemptCompleted
	}
	return a.ToResultDTO(), nil
}
