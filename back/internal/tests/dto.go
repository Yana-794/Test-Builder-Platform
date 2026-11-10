package tests

import "time"

type CreateTestRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Subject     *string `json:"subject"`
}

// PUT /tests/:id/draft.
type SaveDraftRequest struct {
	ExpectedRevision int         `json:"expectedRevision"`
	Test             DraftTest   `json:"test"`
	Questions        []DraftItem `json:"questions"`
}

type DraftTest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Subject     *string `json:"subject"`
}

// вопрос в черновике
// ID == nil для нового вопроса; для существующего — UUID строкой
type DraftItem struct {
	ID                *string           `json:"id"`
	Prompt            string            `json:"prompt"`
	Type              string            `json:"type"`
	Points            float64           `json:"points"`
	Options           []DraftOptionItem `json:"options"`
	CorrectTextAnswer *string           `json:"correctTextAnswer"`
}

type DraftOptionItem struct {
	ID        *string `json:"id"`
	Text      string  `json:"text"`
	IsCorrect bool    `json:"isCorrect"`
}

// Ответы

type TestResponse struct {
	Test TestDTO `json:"test"`
}

// GET /tests/:id и ответ PUT /draft
// Содержит вопросы и правильные ответы (только владельцу)
type TestEditorResponse struct {
	Test EditorTestDTO `json:"test"`
}

type TestDTO struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Subject     *string `json:"subject"`
	Status      string  `json:"status"`
	Revision    int     `json:"revision"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type EditorTestDTO struct {
	TestDTO
	Questions []QuestionDTO `json:"questions"`
}

type QuestionDTO struct {
	ID                string            `json:"id"`
	Prompt            string            `json:"prompt"`
	Type              string            `json:"type"`
	Points            float64           `json:"points"`
	Position          int               `json:"position"`
	Options           []AnswerOptionDTO `json:"options,omitempty"`
	CorrectTextAnswer *string           `json:"correctTextAnswer,omitempty"`
}

type AnswerOptionDTO struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
	Position  int    `json:"position"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldDetail `json:"details,omitempty"`
}

type FieldDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func toTestDTO(t Test) TestDTO {
	return TestDTO{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Subject:     t.Subject,
		Status:      string(t.Status),
		Revision:    t.Revision,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toEditorTestDTO(t Test, questions []Question) EditorTestDTO {
	qdtos := make([]QuestionDTO, len(questions))
	for i, q := range questions {
		qdtos[i] = toQuestionDTO(q)
	}
	return EditorTestDTO{
		TestDTO:   toTestDTO(t),
		Questions: qdtos,
	}
}

func toQuestionDTO(q Question) QuestionDTO {
	var opts []AnswerOptionDTO
	if len(q.Options) > 0 {
		opts = make([]AnswerOptionDTO, len(q.Options))
		for i, o := range q.Options {
			opts[i] = AnswerOptionDTO{
				ID:        o.ID.String(),
				Text:      o.Text,
				IsCorrect: o.IsCorrect,
				Position:  o.Position,
			}
		}
	}
	return QuestionDTO{
		ID:                q.ID.String(),
		Prompt:            q.Prompt,
		Type:              string(q.Type),
		Points:            q.Points,
		Position:          q.Position,
		Options:           opts,
		CorrectTextAnswer: q.CorrectTextAnswer,
	}
}

type TestListResponse struct {
	Tests  []TestListItemDTO `json:"tests"`
	Filter string            `json:"filter"`
	Total  int               `json:"total"`
}

type TestListItemDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Revision  int    `json:"revision"`
	UpdatedAt string `json:"updatedAt"`
}

func toTestListItemDTO(t Test) TestListItemDTO {
	return TestListItemDTO{
		ID:        t.ID.String(),
		Title:     t.Title,
		Status:    string(t.Status),
		Revision:  t.Revision,
		UpdatedAt: t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}