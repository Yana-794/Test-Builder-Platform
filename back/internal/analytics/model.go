package analytics

import (
	"time"

	"github.com/google/uuid"
)

type RecentAttempt struct {
	AttemptID       uuid.UUID
	StudentName     string
	TestID          uuid.UUID
	TestTitle       string
	Status          string
	Score           *float64
	MaxScore        *float64
	Percent         *float64
	DurationSeconds *int
	StartedAt       time.Time
	CompletedAt     *time.Time
}

type RecentAttemptDTO struct {
	AttemptID       string   `json:"attemptId"`
	StudentName     string   `json:"studentName"`
	TestID          string   `json:"testId"`
	TestTitle       string   `json:"testTitle"`
	Status          string   `json:"status"`
	Score           *float64 `json:"score,omitempty"`
	MaxScore        *float64 `json:"maxScore,omitempty"`
	Percent         *float64 `json:"percent,omitempty"`
	DurationSeconds *int     `json:"durationSeconds,omitempty"`
	StartedAt       string   `json:"startedAt"`
	CompletedAt     *string  `json:"completedAt,omitempty"`
}

type RecentResponse struct {
	Attempts []RecentAttemptDTO `json:"attempts"`
}


func (a RecentAttempt) ToDTO() RecentAttemptDTO {
	dto := RecentAttemptDTO{
		AttemptID:       a.AttemptID.String(),
		StudentName:     a.StudentName,
		TestID:          a.TestID.String(),
		TestTitle:       a.TestTitle,
		Status:          a.Status,
		Score:           a.Score,
		MaxScore:        a.MaxScore,
		Percent:         a.Percent,
		DurationSeconds: a.DurationSeconds,
		StartedAt:       a.StartedAt.UTC().Format(time.RFC3339),
	}
	if a.CompletedAt != nil {
		s := a.CompletedAt.UTC().Format(time.RFC3339)
		dto.CompletedAt = &s
	}
	return dto
}