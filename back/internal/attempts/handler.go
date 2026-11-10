package attempts

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"test-builder/internal/publications"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ctxAttemptKey = "attempt"

type attemptsService interface {
	GetPublicTest(ctx context.Context, token string) (publications.PublicTest, error)
	CreateAttempt(ctx context.Context, token, studentName string) (Attempt, string, error)
	AuthenticateAttempt(ctx context.Context, attemptID uuid.UUID, secret string) (Attempt, error)
	GetState(ctx context.Context, attempt Attempt) (AttemptStateResponse, error)
	SaveAnswers(ctx context.Context, attempt Attempt, req SaveAnswersRequest) error
	Submit(ctx context.Context, attempt Attempt) (AttemptResultDTO, error)
	GetResult(ctx context.Context, attempt Attempt) (AttemptResultDTO, error)
}

type Handler struct {
	service      attemptsService
	secureCookie bool
}

func NewHandler(service attemptsService, appEnv string) *Handler {
	return &Handler{service: service, secureCookie: appEnv == "production"}
}

func writeError(c *gin.Context, status int, code, message string, details ...FieldDetail) {
	body := ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	c.JSON(status, ErrorResponse{Error: body})
}

func respondError(c *gin.Context, err error, op string) {
	var verr *ValidationError
	switch {
	case errors.As(err, &verr):
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", verr.Details...)
	case errors.Is(err, ErrPublicationNotFound), errors.Is(err, ErrAttemptNotFound):
		writeError(c, http.StatusNotFound, "NOT_FOUND", "Not found")
	case errors.Is(err, ErrAttemptExpired):
		writeError(c, http.StatusConflict, "CONFLICT", "Attempt expired")
	case errors.Is(err, ErrAttemptCompleted):
		writeError(c, http.StatusConflict, "CONFLICT", "Attempt already completed")
	default:
		slog.ErrorContext(c.Request.Context(), op+" failed", "error", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
	}
}


// проверяет cookie и кладёт попытку в контекст
func (h *Handler) RequireAttemptCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")

		attemptID, err := uuid.Parse(c.Param("attemptId"))
		if err != nil {
			writeError(c, http.StatusNotFound, "NOT_FOUND", "Attempt not found")
			c.Abort()
			return
		}
		secret, err := readAttemptCookie(c, attemptID)
		if err != nil || secret == "" {
			writeError(c, http.StatusNotFound, "NOT_FOUND", "Attempt not found")
			c.Abort()
			return
		}
		attempt, err := h.service.AuthenticateAttempt(c.Request.Context(), attemptID, secret)
		if err != nil {
			writeError(c, http.StatusNotFound, "NOT_FOUND", "Attempt not found")
			c.Abort()
			return
		}
		c.Set(ctxAttemptKey, attempt)
		c.Next()
	}
}

func attemptFromCtx(c *gin.Context) Attempt {
	return c.MustGet(ctxAttemptKey).(Attempt)
}

// GetPublicTest godoc
// @Summary      Публичный тест по токену
// @Tags         public
// @Produce      json
// @Param        token path string true "Публичный токен"
// @Success      200  {object}  map[string]publications.PublicTest
// @Failure      404  {object}  ErrorResponse
// @Router       /public/tests/{token} [get]
func (h *Handler) GetPublicTest(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	test, err := h.service.GetPublicTest(c.Request.Context(), c.Param("token"))
	if err != nil {
		respondError(c, err, "get public test")
		return
	}
	c.JSON(http.StatusOK, gin.H{"test": test})
}

// CreateAttempt godoc
// @Summary      Создать попытку прохождения
// @Tags         public
// @Accept       json
// @Produce      json
// @Param        token   path string                true "Публичный токен"
// @Param        request body CreateAttemptRequest  true "Имя ученика"
// @Success      201  {object}  CreateAttemptResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /public/tests/{token}/attempts [post]
func (h *Handler) CreateAttempt(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	var req CreateAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	attempt, secret, err := h.service.CreateAttempt(c.Request.Context(), c.Param("token"), req.StudentName)
	if err != nil {
		respondError(c, err, "create attempt")
		return
	}

	setAttemptCookie(c, attempt.ID, secret, attempt.ExpiresAt, h.secureCookie)

	c.JSON(http.StatusCreated, CreateAttemptResponse{Attempt: AttemptShortDTO{
		ID:        attempt.ID.String(),
		Status:    attempt.Status,
		StartedAt: attempt.StartedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}})
}

// GetState godoc
// @Summary      Состояние попытки
// @Tags         public
// @Produce      json
// @Param        attemptId path string true "UUID попытки"
// @Success      200  {object}  AttemptStateResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /public/attempts/{attemptId} [get]
func (h *Handler) GetState(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	state, err := h.service.GetState(c.Request.Context(), attemptFromCtx(c))
	if err != nil {
		respondError(c, err, "get state")
		return
	}
	c.JSON(http.StatusOK, state)
}

// SaveAnswers godoc
// @Summary      Сохранить ответы
// @Tags         public
// @Accept       json
// @Param        attemptId path string             true "UUID попытки"
// @Param        request   body SaveAnswersRequest true "Полный набор ответов"
// @Success      204
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /public/attempts/{attemptId}/answers [put]
func (h *Handler) SaveAnswers(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	var req SaveAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	if err := h.service.SaveAnswers(c.Request.Context(), attemptFromCtx(c), req); err != nil {
		respondError(c, err, "save answers")
		return
	}
	c.Status(http.StatusNoContent)
}

// Submit godoc
// @Summary      Завершить попытку
// @Tags         public
// @Produce      json
// @Param        attemptId path string true "UUID попытки"
// @Success      200  {object}  AttemptResultResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /public/attempts/{attemptId}/submit [post]
func (h *Handler) Submit(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	result, err := h.service.Submit(c.Request.Context(), attemptFromCtx(c))
	if err != nil {
		respondError(c, err, "submit")
		return
	}
	c.JSON(http.StatusOK, AttemptResultResponse{Attempt: result})
}

// GetResult godoc
// @Summary      Результат попытки
// @Tags         public
// @Produce      json
// @Param        attemptId path string true "UUID попытки"
// @Success      200  {object}  AttemptResultResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Router       /public/attempts/{attemptId}/result [get]
func (h *Handler) GetResult(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	result, err := h.service.GetResult(c.Request.Context(), attemptFromCtx(c))
	if err != nil {
		respondError(c, err, "get result")
		return
	}
	c.JSON(http.StatusOK, AttemptResultResponse{Attempt: result})
}