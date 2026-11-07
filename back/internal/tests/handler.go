package tests

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"test-builder/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
type testsService interface {
	CreateTest(ctx context.Context, ownerID uuid.UUID, req CreateTestRequest) (Test, error)
	GetTestForEditor(ctx context.Context, ownerID, testID uuid.UUID) (Test, []Question, error)
	SaveDraft(ctx context.Context, ownerID, testID uuid.UUID, req SaveDraftRequest) (Test, []Question, error)
}
type Handler struct {
	service testsService}

func NewHandler(service testsService) *Handler {
	return &Handler{service: service}
}

func writeError(c *gin.Context, status int, code, message string, details []FieldDetail) {
	body := ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	c.JSON(status, ErrorResponse{Error: body})
}

func bindError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeError(c, http.StatusRequestEntityTooLarge,
			"REQUEST_TOO_LARGE", "Request body is too large", nil)
		return
	}
	writeError(c, http.StatusBadRequest,
		"VALIDATION_ERROR", "Invalid request body", nil)
}

// POST /tests

func (h *Handler) CreateTest(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	userID, ok := auth.UserIDFromCtx(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Please sign in", nil)
		return
	}

	var req CreateTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	test, err := h.service.CreateTest(c.Request.Context(), userID, req)
	if err != nil {
		var verr *ValidationError
		if errors.As(err, &verr) {
			writeError(c, http.StatusBadRequest,
				"VALIDATION_ERROR", "Request validation failed", verr.Details)
			return
		}
		slog.ErrorContext(c.Request.Context(), "create test failed", "error", err)
		writeError(c, http.StatusInternalServerError,
			"INTERNAL_ERROR", "Something went wrong", nil)
		return
	}

	c.JSON(http.StatusCreated, TestResponse{Test: toTestDTO(test)})
}

// GET /tests/:testId

func (h *Handler) GetTest(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	userID, ok := auth.UserIDFromCtx(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Please sign in", nil)
		return
	}

	testID, err := uuid.Parse(c.Param("testId"))
	if err != nil {
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found", nil)
		return
	}

	test, questions, err := h.service.GetTestForEditor(c.Request.Context(), userID, testID)
	if err != nil {
		if errors.Is(err, ErrTestNotFound) {
			writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found", nil)
			return
		}
		slog.ErrorContext(c.Request.Context(), "get test failed", "error", err)
		writeError(c, http.StatusInternalServerError,
			"INTERNAL_ERROR", "Something went wrong", nil)
		return
	}

	c.JSON(http.StatusOK, TestEditorResponse{Test: toEditorTestDTO(test, questions)})
}

// PUT /tests/:testId/draft

func (h *Handler) SaveDraft(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	userID, ok := auth.UserIDFromCtx(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Please sign in", nil)
		return
	}

	testID, err := uuid.Parse(c.Param("testId"))
	if err != nil {
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found", nil)
		return
	}

	var req SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	test, questions, err := h.service.SaveDraft(c.Request.Context(), userID, testID, req)
	if err != nil {
		var verr *ValidationError
		switch {
		case errors.As(err, &verr):
			writeError(c, http.StatusBadRequest,
				"VALIDATION_ERROR", "Request validation failed", verr.Details)
		case errors.Is(err, ErrTestNotFound):
			writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found", nil)
		case errors.Is(err, ErrTestNotDraft):
			writeError(c, http.StatusConflict,
				"CONFLICT", "Test is not a draft", nil)
		case errors.Is(err, ErrRevisionConflict):
			writeError(c, http.StatusConflict,
				"CONFLICT", "Draft was modified in another tab. Refresh to continue.", nil)
		case errors.Is(err, ErrInvalidReference):
			writeError(c, http.StatusBadRequest,
				"VALIDATION_ERROR", err.Error(), nil)
		default:
			slog.ErrorContext(c.Request.Context(), "save draft failed", "error", err)
			writeError(c, http.StatusInternalServerError,
				"INTERNAL_ERROR", "Something went wrong", nil)
		}
		return
	}

	c.JSON(http.StatusOK, TestEditorResponse{Test: toEditorTestDTO(test, questions)})
}