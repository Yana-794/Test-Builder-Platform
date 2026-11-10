package publications

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusDraft      = "draft"
	TypeSingleChoice = "single_choice"
	TypeText         = "text"
)
const codeRequired = "REQUIRED"
const (
	maxPromptLen        = 2000
	minSingleChoiceOpts = 2
	snapshotSchemaVer   = 1
)

type Service struct {
	repo          *Repository
	pool          *pgxpool.Pool
	publicBaseURL string
}

func NewService(repo *Repository, pool *pgxpool.Pool, publicBaseURL string) *Service {
	return &Service{
		repo:          repo,
		pool:          pool,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (s *Service) Validate(
	ctx context.Context,
	ownerID, testID uuid.UUID,
) (ValidateResponse, error) {
	draft, err := s.repo.LoadDraftForValidation(ctx, s.pool, ownerID, testID)
	if err != nil {
		return ValidateResponse{}, err
	}

	response := validateDraft(draft)
	response.Summary = ValidateSummary{
		QuestionCount: len(draft.Questions),
		MaxScore:      sumPoints(draft.Questions),
	}
	return response, nil
}

func validateDraft(draft DraftForPublish) ValidateResponse {
	if len(draft.Questions) == 0 {
		response := ValidateResponse{Errors: []FieldDetail{{
			Field:   "questions",
			Code:    codeRequired,
			Message: "Тест должен содержать минимум один вопрос",
		}}}
		response.Valid = false
		return response
	}

	var details []FieldDetail
	for questionIndex, question := range draft.Questions {
		details = append(details, validateQuestion(questionIndex, question)...)
	}
	if sumPoints(draft.Questions) <= 0 {
		details = append(details, FieldDetail{
			Field:   "questions",
			Code:    "INVALID_MAX_SCORE",
			Message: "Сумма баллов должна быть больше нуля",
		})
	}

	response := ValidateResponse{Errors: details}
	response.Valid = len(details) == 0
	return response
}

func validateQuestion(questionIndex int, question QuestionRow) []FieldDetail {
	field := fmt.Sprintf("questions[%d]", questionIndex)
	var details []FieldDetail

	if strings.TrimSpace(question.Prompt) == "" {
		details = append(details, FieldDetail{
			Field:   field + ".prompt",
			Code:    codeRequired,
			Message: "Поле 'Вопрос' обязательно для заполнения",
		})
	} else if utf8.RuneCountInString(question.Prompt) > maxPromptLen {
		details = append(details, FieldDetail{
			Field:   field + ".prompt",
			Code:    "TOO_LONG",
			Message: fmt.Sprintf("Поле 'Вопрос' не должно превышать %d символов", maxPromptLen),
		})
	}

	if question.Points <= 0 {
		details = append(details, FieldDetail{
			Field:   field + ".points",
			Code:    "INVALID_POINTS",
			Message: "Баллы за вопрос должны быть больше нуля",
		})
	}

	switch question.Type {
	case TypeSingleChoice:
		details = append(details, validateSingleChoice(field, question)...)
	case TypeText:
		details = append(details, validateText(field, question)...)
	default:
		details = append(details, FieldDetail{
			Field:   field + ".type",
			Code:    "INVALID_TYPE",
			Message: "Недопустимый тип вопроса",
		})
	}
	return details
}

func validateSingleChoice(field string, question QuestionRow) []FieldDetail {
	filled, correct := 0, 0
	for _, option := range question.Options {
		if strings.TrimSpace(option.Text) == "" {
			continue
		}
		filled++
		if option.IsCorrect {
			correct++
		}
	}

	var details []FieldDetail
	if filled < minSingleChoiceOpts {
		details = append(details, FieldDetail{
			Field:   field + ".options",
			Code:    "NOT_ENOUGH_OPTIONS",
			Message: fmt.Sprintf("Вопрос должен содержать минимум %d варианта ответа", minSingleChoiceOpts),
		})
	}
	if correct != 1 {
		details = append(details, FieldDetail{
			Field:   field + ".options",
			Code:    "INVALID_CORRECT_OPTIONS",
			Message: "Должен быть ровно один правильный вариант",
		})
	}
	return details
}

func validateText(field string, question QuestionRow) []FieldDetail {
	if question.CorrectTextAnswer == nil || strings.TrimSpace(*question.CorrectTextAnswer) == "" {
		return []FieldDetail{{
			Field:   field + ".correctTextAnswer",
			Code:    codeRequired,
			Message: "Правильный ответ обязателен для текстового вопроса",
		}}
	}
	return nil
}

func sumPoints(questions []QuestionRow) float64 {
	var total float64
	for _, question := range questions {
		total += question.Points
	}
	return total
}

func (s *Service) Publish(
	ctx context.Context,
	ownerID, testID uuid.UUID,
	expectedRevision int,
) (PublishResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PublishResponse{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	draft, err := s.repo.LoadDraftForPublish(ctx, tx, ownerID, testID)
	if err != nil {
		return PublishResponse{}, err
	}

	if draft.Test.Status != StatusDraft {
		return PublishResponse{}, ErrTestNotDraft
	}
	if draft.Test.Revision != expectedRevision {
		return PublishResponse{}, ErrRevisionConflict
	}
	validationResult := validateDraft(draft)
	if !validationResult.Valid {
		return PublishResponse{}, &ValidationError{Details: validationResult.Errors}
	}

	version, err := s.repo.NextVersion(ctx, tx, testID)
	if err != nil {
		return PublishResponse{}, err
	}

	token, err := generatePublicToken()
	if err != nil {
		return PublishResponse{}, fmt.Errorf("generate token: %w", err)
	}

	pub, err := s.repo.CreatePublication(ctx, tx, testID, version, token, buildSnapshot(draft))
	if err != nil {
		return PublishResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return PublishResponse{}, fmt.Errorf("commit: %w", err)
	}

	return PublishResponse{
		Publication: PublishResponseData{
			ID:          pub.ID.String(),
			Version:     pub.Version,
			PublicToken: pub.PublicToken,
			PublicURL:   s.PublicURL(pub.PublicToken),
			PublishedAt: pub.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func generatePublicToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func buildSnapshot(draft DraftForPublish) SnapshotContent {
	qsnapshot := make([]SnapshotQuestion, len(draft.Questions))

	for questionIndex, question := range draft.Questions {
		snapshotQuestion := SnapshotQuestion{
			ID:                question.ID,
			Prompt:            question.Prompt,
			Type:              question.Type,
			Points:            question.Points,
			Position:          question.Position,
			CorrectTextAnswer: question.CorrectTextAnswer,
		}
		if len(question.Options) > 0 {
			snapshotQuestion.Options = make([]SnapshotOption, len(question.Options))
			for optionIndex, option := range question.Options {
				snapshotQuestion.Options[optionIndex] = SnapshotOption{
					ID:        option.ID,
					Text:      option.Text,
					IsCorrect: option.IsCorrect,
					Position:  option.Position,
				}
			}
		}
		qsnapshot[questionIndex] = snapshotQuestion
	}

	return SnapshotContent{
		SchemaVersion: snapshotSchemaVer,
		Title:         draft.Test.Title,
		Description:   draft.Test.Description,
		Subject:       draft.Test.Subject,
		Questions:     qsnapshot,
	}
}

func (s *Service) PublicURL(token string) string {
	publicURL, err := url.JoinPath(s.publicBaseURL, "test", token)
	if err != nil {
		return ""
	}
	return publicURL
}
