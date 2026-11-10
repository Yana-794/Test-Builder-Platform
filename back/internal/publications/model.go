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
	Title         string
	Description   string
	QuestionCount int
	Questions     []PublicQuestion
}

type PublicQuestion struct {
	ID       uuid.UUID
	Prompt   string
	Type     string
	Position int
	Options  []PublicOption 
}

//только для одиночного выбора
type PublicOption struct {
	ID   uuid.UUID
	Text string
}