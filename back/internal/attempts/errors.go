package attempts

import (
	"errors"
	"test-builder/internal/publications"
)

var (
	ErrPublicationNotFound = errors.New("publication not found")
	ErrAttemptNotFound     = errors.New("attempt not found")
	ErrAttemptExpired      = errors.New("attempt expired")
	ErrAttemptCompleted    = errors.New("attempt already completed")
)

type (
	ValidationError = publications.ValidationError
	FieldDetail     = publications.FieldDetail
	ErrorResponse   = publications.ErrorResponse
	ErrorBody       = publications.ErrorBody
)