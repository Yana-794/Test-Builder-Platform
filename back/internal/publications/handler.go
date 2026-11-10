package publications

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"test-builder/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
type publicationsService interface {
	Validate(ctx context.Context, ownerID, testID uuid.UUID) (ValidateResponse, error)
	Publish(ctx context.Context, ownerID, testID uuid.UUID, expectedRevision int) (PublishResponse, error)
}

type Handler struct {
	service publicationsService  
}

func NewHandler(service publicationsService) *Handler {
	return &Handler{service: service}
}

func writeError(c *gin.Context, status int, code, message string, details ...FieldDetail) {
	body := ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	c.JSON(status, ErrorResponse{Error: body})
}

func requireAuthAndTestID(c *gin.Context) (userID, testID uuid.UUID, ok bool) {
	userID, ok = auth.UserIDFromCtx(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Please sign in")
		return uuid.Nil, uuid.Nil, false
	}

	testID, err := uuid.Parse(c.Param("testId"))
	if err != nil {
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found")
		return uuid.Nil, uuid.Nil, false
	}

	return userID, testID, true
}

// единый маппинг доменных ошибок в HTTP-коды
func respondError(c *gin.Context, err error, operation string) {
	var verr *ValidationError
	if errors.As(err, &verr) {
		writeError(c, http.StatusUnprocessableEntity,"PUBLISH_VALIDATION_FAILED", "Test is not ready for publishing", verr.Details...)
		return
	}

	switch {
	case errors.Is(err, ErrTestNotFound):
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Test not found")
	case errors.Is(err, ErrTestNotDraft):
		writeError(c, http.StatusConflict, "CONFLICT", "Test is not a draft")
	case errors.Is(err, ErrRevisionConflict):
		writeError(c, http.StatusConflict, "CONFLICT",
			"Draft was modified in another tab. Refresh to continue.")
	case errors.Is(err, ErrNotReady):
		writeError(c, http.StatusUnprocessableEntity,
			"PUBLISH_VALIDATION_FAILED", "Test is not ready for publishing")
	default:
		slog.ErrorContext(c.Request.Context(), operation+" failed", "error", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
	}
}

// Validate godoc
// @Summary      Проверка готовности теста к публикации
// @Tags         publications
// @Produce      json
// @Param        testId path string true "UUID теста"
// @Success      200  {object}  ValidateResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tests/{testId}/validate [post]
// @Security     CookieAuth
// @Security     CsrfToken
func (h *Handler) Validate(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, testID, ok := requireAuthAndTestID(c)
	if !ok {
		return
	}
	resp, err := h.service.Validate(c.Request.Context(), userID, testID)
	if err != nil {
		respondError(c, err, "validate")
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Publish godoc
// @Summary      Публикация теста
// @Tags         publications
// @Accept       json
// @Produce      json
// @Param        testId  path string          true "UUID теста"
// @Param        request body PublishRequest  true "expectedRevision"
// @Success      201  {object}  PublishResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      422  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tests/{testId}/publish [post]
// @Security     CookieAuth
// @Security     CsrfToken
func (h *Handler) Publish(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, testID, ok := requireAuthAndTestID(c)
	if !ok {
		return
	}

	var req PublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}
	if req.ExpectedRevision < 1 {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "expectedRevision must be >= 1")
		return
	}

	pub, err := h.service.Publish(c.Request.Context(), userID, testID, req.ExpectedRevision)
	if err != nil {
		respondError(c, err, "publish")
		return
	}

	c.JSON(http.StatusCreated, pub)
}
