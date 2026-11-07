package tests

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateTest

func (s *Service) CreateTest(
	ctx context.Context,
	ownerID uuid.UUID,
	req CreateTestRequest,
) (Test, error) {
	var details []FieldDetail

	title := strings.TrimSpace(req.Title)
	if title == "" {
		details = append(details, FieldDetail{"title", "is required"})
	} else if utf8.RuneCountInString(title) > MaxTitleLen {
		details = append(details, FieldDetail{
			"title", fmt.Sprintf("must be at most %d characters", MaxTitleLen),
		})
	}

	description := strings.TrimSpace(req.Description)
	if utf8.RuneCountInString(description) > MaxDescriptionLen {
		details = append(details, FieldDetail{
			"description", fmt.Sprintf("must be at most %d characters", MaxDescriptionLen),
		})
	}

	subject := normalizeOptionalString(req.Subject)
	if subject != nil && utf8.RuneCountInString(*subject) > MaxSubjectLen {
		details = append(details, FieldDetail{
			"subject", fmt.Sprintf("must be at most %d characters", MaxSubjectLen),
		})
	}

	if len(details) > 0 {
		return Test{}, &ValidationError{Details: details}
	}

	return s.repo.CreateTest(ctx, ownerID, title, description, subject)
}

// GetTestForEditor

func (s *Service) GetTestForEditor(
	ctx context.Context,
	ownerID, testID uuid.UUID,
) (Test, []Question, error) {
	return s.repo.GetTestWithQuestions(ctx, ownerID, testID)
}

// SaveDraft

func (s *Service) SaveDraft(
	ctx context.Context,
	ownerID, testID uuid.UUID,
	req SaveDraftRequest,
) (Test, []Question, error) {
	if err := validateSaveDraft(req); err != nil {
		return Test{}, nil, err
	}
	normalizeSaveDraft(&req)
	return s.repo.SaveDraft(ctx, ownerID, testID, req.ExpectedRevision, req)
}

func validateSaveDraft(req SaveDraftRequest) error {
	var details []FieldDetail

	if req.ExpectedRevision < 1 {
		details = append(details, FieldDetail{"expectedRevision", "must be >= 1"})
	}

	title := strings.TrimSpace(req.Test.Title)
	if title == "" {
		details = append(details, FieldDetail{"test.title", "is required"})
	} else if utf8.RuneCountInString(title) > MaxTitleLen {
		details = append(details, FieldDetail{
			"test.title", fmt.Sprintf("must be at most %d characters", MaxTitleLen),
		})
	}

	if utf8.RuneCountInString(strings.TrimSpace(req.Test.Description)) > MaxDescriptionLen {
		details = append(details, FieldDetail{
			"test.description", fmt.Sprintf("must be at most %d characters", MaxDescriptionLen),
		})
	}

	if req.Test.Subject != nil {
		if utf8.RuneCountInString(strings.TrimSpace(*req.Test.Subject)) > MaxSubjectLen {
			details = append(details, FieldDetail{
				"test.subject", fmt.Sprintf("must be at most %d characters", MaxSubjectLen),
			})
		}
	}

	if len(req.Questions) > MaxQuestionsPerTest {
		details = append(details, FieldDetail{
			"questions", fmt.Sprintf("must have at most %d items", MaxQuestionsPerTest),
		})
	}

	for i, q := range req.Questions {
		field := fmt.Sprintf("questions[%d]", i)

		if utf8.RuneCountInString(strings.TrimSpace(q.Prompt)) > MaxPromptLen {
			details = append(details, FieldDetail{
				field + ".prompt", fmt.Sprintf("must be at most %d characters", MaxPromptLen),
			})
		}

		if !QuestionType(q.Type).Valid() {
			details = append(details, FieldDetail{
				field + ".type", "must be 'single_choice' or 'text'",
			})
			continue // дальше проверки зависят от типа
		}

		details = append(details, validatePoints(field+".points", q.Points)...)

		switch q.Type {
		case QuestionTypeSingleChoice:
			if q.CorrectTextAnswer != nil && strings.TrimSpace(*q.CorrectTextAnswer) != "" {
				details = append(details, FieldDetail{
					field + ".correctTextAnswer", "must be empty for single_choice",
				})
			}
			details = append(details, validateSingleChoiceOptions(field, q.Options)...)

		case QuestionTypeText:
			if len(q.Options) > 0 {
				details = append(details, FieldDetail{
					field + ".options", "must be empty for text question",
				})
			}
			if q.CorrectTextAnswer != nil {
				if utf8.RuneCountInString(*q.CorrectTextAnswer) > MaxCorrectTextAnswerLen {
					details = append(details, FieldDetail{
						field + ".correctTextAnswer",
						fmt.Sprintf("must be at most %d characters", MaxCorrectTextAnswerLen),
					})
				}
			}
		}
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}
	return nil
}

func validatePoints(field string, points float64) []FieldDetail {
	if points <= 0 {
		return []FieldDetail{{field, "must be positive"}}
	}
	if points > 9999999.99 {
		return []FieldDetail{{field, "too large"}}
	}
	// Точность до 2 знаков
	if math.Abs(points*100-math.Round(points*100)) > 1e-6 {
		return []FieldDetail{{field, "must have at most 2 decimal places"}}
	}
	return nil
}

func validateSingleChoiceOptions(field string, opts []DraftOptionItem) []FieldDetail {
	var details []FieldDetail

	if len(opts) > MaxOptionsPerQuestion {
		details = append(details, FieldDetail{
			field + ".options", fmt.Sprintf("must have at most %d items", MaxOptionsPerQuestion),
		})
	}

	correctCount := 0
	for j, opt := range opts {
		optField := fmt.Sprintf("%s.options[%d]", field, j)
		if utf8.RuneCountInString(strings.TrimSpace(opt.Text)) > MaxOptionTextLen {
			details = append(details, FieldDetail{
				optField + ".text", fmt.Sprintf("must be at most %d characters", MaxOptionTextLen),
			})
		}
		if opt.IsCorrect {
			correctCount++
		}
	}
	if correctCount > 1 {
		details = append(details, FieldDetail{
			field + ".options", "only one option may be marked as correct",
		})
	}
	return details
}

// приводит все текстовые поля к чистому виду перед передачей в репозиторий
// модифицирует req на месте
func normalizeSaveDraft(req *SaveDraftRequest) {
	req.Test.Title = strings.TrimSpace(req.Test.Title)
	req.Test.Description = strings.TrimSpace(req.Test.Description)
	req.Test.Subject = normalizeOptionalString(req.Test.Subject)

	for i := range req.Questions {
		q := &req.Questions[i]
		q.Prompt = strings.TrimSpace(q.Prompt)

		if q.Type == QuestionTypeText {
			if q.CorrectTextAnswer != nil {
				s := strings.TrimSpace(*q.CorrectTextAnswer)
				if s == "" {
					q.CorrectTextAnswer = nil
				} else {
					q.CorrectTextAnswer = &s
				}
			}
		} else {
			// правильного текстового ответа быть не может
			q.CorrectTextAnswer = nil
		}

		for j := range q.Options {
			q.Options[j].Text = strings.TrimSpace(q.Options[j].Text)
		}
	}
}

func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
