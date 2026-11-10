package publications

import (
	"time"

	"github.com/google/uuid"
)

type Publication struct {
	ID          uuid.UUID
	TestID      uuid.UUID
	Version     int
	PublicToken string
	Content     SnapshotContent
	CreatedAt    time.Time
}

// сериализуется в JSONB при публикации, читается при прохождении
type SnapshotContent struct {
	SchemaVersion int `json:"version"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Subject     *string            `json:"subject,omitempty"`
	Questions   []SnapshotQuestion `json:"questions"`
}

type SnapshotQuestion struct {
	ID                uuid.UUID        `json:"id"`
	Prompt            string           `json:"prompt"`
	Type              string           `json:"type"` // single_choice | text
	Points            float64          `json:"points"`
	Position          int              `json:"position"`
	Options           []SnapshotOption `json:"options,omitempty"`
	CorrectTextAnswer *string          `json:"correctTextAnswer,omitempty"`
}
 type SnapshotOption struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	IsCorrect bool      `json:"isCorrect"`
	Position  int       `json:"position"`
 }

type PublicTest struct {
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	QuestionCount int              `json:"questionCount"`
	Questions     []PublicQuestion `json:"questions"`
}

type PublicQuestion struct {
	ID       uuid.UUID      `json:"id"`
	Prompt   string         `json:"prompt"`
	Type     string         `json:"type"`
	Position int            `json:"position"`
	Options  []PublicOption `json:"options,omitempty"` 
}

//только для одиночного выбора
type PublicOption struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

func ToPublicTest(s SnapshotContent) PublicTest {
	qs := make([]PublicQuestion, len(s.Questions))
	for i, q := range s.Questions {
		pq := PublicQuestion{ID: q.ID, Prompt: q.Prompt, Type: q.Type, Position: q.Position}
		if q.Type == "single_choice" {
			pq.Options = make([]PublicOption, len(q.Options))
			for j, o := range q.Options {
				pq.Options[j] = PublicOption{ID: o.ID, Text: o.Text}
			}
		}
		qs[i] = pq
	}
	return PublicTest{
		Title:         s.Title,
		Description:   s.Description,
		QuestionCount: len(s.Questions),
		Questions:     qs,
	}
}