package dashboard

import (
	"context"
	"log/slog"
	"net/http"

	"test-builder/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service overviewGetter
}

// Позволяет подставить мок в тестах.
type overviewGetter interface {
	GetOverview(ctx context.Context, userID uuid.UUID) (DashboardOverview, error)
}

func NewHandler(svc overviewGetter) *Handler {
	return &Handler{service: svc}
}

// GetOverview godoc
// @Summary      Сводка Dashboard
// @Description  Статистика тестов и последние 5 тестов текущего пользователя
// @Tags         dashboard
// @Produce      json
// @Success      200  {object}  DashboardResponse
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /dashboard/overview [get]
// @Security     CookieAuth
func (h *Handler) GetOverview(c *gin.Context) {
	 c.Header("Cache-Control", "no-store")

	ownerID, ok := auth.UserIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHENTICATED",
				"message": "Please sign in",
			},
		})
		return
	}

	overview, err := h.service.GetOverview(c.Request.Context(), ownerID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(),
			"dashboard overview failed",
			"owner_id", ownerID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Something went wrong",
			},
		})
		return
	}

	c.JSON(http.StatusOK, ToResponse(overview))
}
