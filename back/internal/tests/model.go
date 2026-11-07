package tests

import (
	"time"

	"github.com/google/uuid"
)

// Используются в SQL-запросах и валидации.

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"

	QuestionTypeSingleChoice = "single_choice"
	QuestionTypeText         = "text"
)



const (
	MaxTitleLen             = 160
	MaxDescriptionLen       = 2000
	MaxSubjectLen           = 100
	MaxPromptLen            = 2000
	MaxOptionTextLen        = 500
	MaxCorrectTextAnswerLen = 1000
	MaxQuestionsPerTest     = 100
	MaxOptionsPerQuestion   = 10
)

type Status string
type QuestionType string

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusArchived:
		return true
	}
	return false
}

func (t QuestionType) Valid() bool {
	switch t {
	case QuestionTypeSingleChoice, QuestionTypeText:
		return true
	}
	return false
}

// Test — внутренняя модель. Не используется как HTTP-ответ.
type Test struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description string
	Subject     *string // nil = не задан
	Status      Status
	Revision    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Question struct {
	ID                uuid.UUID
	TestID            uuid.UUID
	Prompt            string
	Type              QuestionType
	Points            float64
	Position          int
	CorrectTextAnswer *string 
	Options           []AnswerOption
}

type AnswerOption struct {
	ID         uuid.UUID
	QuestionID uuid.UUID
	Text       string
	IsCorrect  bool
	Position   int
}