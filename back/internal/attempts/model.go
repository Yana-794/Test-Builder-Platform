package attempts

import (
	"time"
	"github.com/google/uuid"
)

const (
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	AttemptLifetime = 24 * time.Hour
)

// попытка прохождения теста учеником
// Каждая попытка связана с публикацией (неизменяемым снапшотом)
// Секрет доступа хранится в cookie ученика, в БД — sha256-хеш
type Attempt struct {
	ID              uuid.UUID
	PublicationID   uuid.UUID
	StudentName     string
	Status          string
	AccessTokenHash []byte
	StartedAt       time.Time
	ExpiresAt       time.Time
	CompletedAt     *time.Time
	Score           *float64
	MaxScore        *float64
	ScorePercent    *float64
	DurationSeconds *int
}

// Answer — один ответ ученика на вопрос
type Answer struct {
	QuestionID       uuid.UUID
	AnswerText       *string
	SelectedOptionID *uuid.UUID
}
