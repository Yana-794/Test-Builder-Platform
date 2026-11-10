package attempts

import "test-builder/internal/publications"

// POST /public/tests/:token/attempts
type CreateAttemptRequest struct {
	StudentName string `json:"studentName"`
}

type CreateAttemptResponse struct {
	Attempt AttemptShortDTO `json:"attempt"`
}

type AttemptShortDTO struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	StartedAt string `json:"startedAt"`
}

// GET /public/attempts/:attemptId
type AttemptStateResponse struct {
	Attempt AttemptShortDTO         `json:"attempt"`
	Test    publications.PublicTest `json:"test"`
	Answers []AnswerDTO             `json:"answers"`
}

// PUT /public/attempts/:attemptId/answers

type SaveAnswersRequest struct {
	Answers []AnswerDTO `json:"answers"`
}

//  один ответ ученика
type AnswerDTO struct {
	QuestionID       string  `json:"questionId"`
	SelectedOptionID *string `json:"selectedOptionId,omitempty"`
	AnswerText       *string `json:"answerText,omitempty"`
}

type AttemptResultResponse struct {
	Attempt AttemptResultDTO `json:"attempt"`
}

type AttemptResultDTO struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	Score           float64 `json:"score"`
	MaxScore        float64 `json:"maxScore"`
	Percentage      float64 `json:"percentage"`
	DurationSeconds int     `json:"durationSeconds"`
	CompletedAt     string  `json:"completedAt"`
}

func (a Attempt) ToResultDTO() AttemptResultDTO {
	var score, maxScore, percent float64
	var duration int
	var completedAt string

	if a.Score != nil {
		score = *a.Score
	}
	if a.MaxScore != nil {
		maxScore = *a.MaxScore
	}
	if a.ScorePercent != nil {
		percent = *a.ScorePercent
	}
	if a.DurationSeconds != nil {
		duration = *a.DurationSeconds
	}
	if a.CompletedAt != nil {
		completedAt = a.CompletedAt.UTC().Format("2006-01-02T15:04:05Z")
	}

	return AttemptResultDTO{
		ID:              a.ID.String(),
		Status:          a.Status,
		Score:           score,
		MaxScore:        maxScore,
		Percentage:      percent,
		DurationSeconds: duration,
		CompletedAt:     completedAt,
	}
}
