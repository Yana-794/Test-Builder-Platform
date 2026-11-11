package analytics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"test-builder/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type analyticsService interface {
	Recent(ctx context.Context, ownerID uuid.UUID, limit int) ([]RecentAttemptDTO, error)
}

type Handler struct {
	service analyticsService
}

func NewHandler(service analyticsService) *Handler {
	return &Handler{service: service}
}
func writeError(c *gin.Context, status int, code, message string, details ...FieldDetail) {
	body := ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details
	}
	c.JSON(status, ErrorResponse{Error: body})
}
func requireUserID(c *gin.Context) (uuid.UUID, bool) {
userID, ok := auth.UserIDFromCtx(c)
if !ok {
	writeError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Please sign in")
	return uuid.Nil, false
}
 return  userID, true
}

// Recent godoc
// @Summary      Последние попытки учеников
// @Tags         analytics
// @Produce      json
// @Param        limit query int false "1-100, default 20"
// @Success      200  {object}  RecentResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /analytics/recent [get]
// @Security     CookieAuth
func (h *Handler) Recent(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	limit := 20
	if v := c.Query("limit"); v!= "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <=100 {
			limit= n
		}
	}
	attempts, err := h.service.Recent(c.Request.Context(), userID, limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "analytics recent failed", "error", err)
		writeError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
		return
	}
	c.JSON(http.StatusOK, RecentResponse{Attempts: attempts})
}