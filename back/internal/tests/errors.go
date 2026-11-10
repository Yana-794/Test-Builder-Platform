package tests

import "errors"

var (
	//тест не существует или принадлежит другому пользователю.
	ErrTestNotFound = errors.New("test not found")

	//тест опубликован или архивирован, редактирование запрещено.
	ErrTestNotDraft = errors.New("test is not a draft")

	// expectedRevision не совпадает с текущей revision.
	ErrRevisionConflict = errors.New("revision conflict")

	// передан id вопроса/варианта, который не принадлежит этому тесту/вопросу.
	ErrInvalidReference = errors.New("invalid reference")
	ErrInvalidFilter = errors.New("invalid filter")
	
)
type ValidationError struct {
	Details []FieldDetail
}

func (e *ValidationError) Error() string { return "validation failed" }
