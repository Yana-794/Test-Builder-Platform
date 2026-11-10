package publications

import "errors"

var (
    ErrTestNotFound     = errors.New("test not found")
    ErrTestNotDraft     = errors.New("test is not a draft")
    ErrRevisionConflict = errors.New("revision conflict")
    ErrNotReady         = errors.New("test is not ready for publishing")
)

type ValidationError struct {
    Details []FieldDetail
}

func (e *ValidationError) Error() string { return "validation failed" }

type FieldDetail struct {
    Field   string `json:"field"`
    Code    string `json:"code"`     
    Message string `json:"message"`
}