package dashboard_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"test-builder/internal/auth"
	"test-builder/internal/dashboard"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockService реализует dashboard.overviewGetter.
// Возвращает значение (не указатель) — как реальный *dashboard.Service.
type mockService struct {
	overview dashboard.DashboardOverview
	err      error
}

func (m *mockService) GetOverview(_ context.Context, _ uuid.UUID) (dashboard.DashboardOverview, error) {
	return m.overview, m.err
}

func init() {
	gin.SetMode(gin.TestMode)
}

// withSession кладёт сессию в контекст, имитируя middleware RequireAuth.
func withSession(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authenticatedSession", auth.AuthenticatedSession{
			User: auth.User{ID: uuid.New()},
		})
		h(c)
	}
}

func TestHandler_GetOverview_Success(t *testing.T) {
	mock := &mockService{
		overview: dashboard.DashboardOverview{
			Stats: dashboard.DashboardStats{
				TotalTests:     12,
				DraftTests:     4,
				PublishedTests: 8,
			},
			RecentTests: []dashboard.RecentTest{
				{
					ID:        uuid.New(),
					Title:     "Основы JavaScript",
					Status:    "published",
					CreatedAt: time.Now().UTC(),
				},
			},
		},
	}

	r := gin.New()
	r.GET("/overview", withSession(dashboard.NewHandler(mock).GetOverview))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/overview", nil))

	// 1. статус
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. заголовок — персональные данные не кэшируются
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	// 3. тело — контракт JSON
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	stats, ok := body["stats"].(map[string]interface{})
	require.True(t, ok, "stats должен быть объектом")
	assert.Equal(t, float64(12), stats["totalTests"])
	assert.Equal(t, float64(4), stats["draftTests"])
	assert.Equal(t, float64(8), stats["publishedTests"])

	recent, ok := body["recentTests"].([]interface{})
	require.True(t, ok, "recentTests должен быть массивом, а не null")
	require.Len(t, recent, 1)

	first := recent[0].(map[string]interface{})
	assert.Equal(t, "Основы JavaScript", first["title"])
	assert.Equal(t, "published", first["status"])
}

func TestHandler_GetOverview_Unauthorized(t *testing.T) {
	// сессия не кладётся в контекст → handler должен вернуть 401
	r := gin.New()
	r.GET("/overview", dashboard.NewHandler(&mockService{}).GetOverview)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/overview", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	errObj, ok := body["error"].(map[string]interface{})
	require.True(t, ok, "тело должно содержать error-объект")
	assert.Equal(t, "UNAUTHENTICATED", errObj["code"])
}

func TestHandler_GetOverview_InternalError(t *testing.T) {
	mock := &mockService{err: errors.New("db down")}

	r := gin.New()
	r.GET("/overview", withSession(dashboard.NewHandler(mock).GetOverview))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/overview", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	errObj, ok := body["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", errObj["code"])

	// важно: детали ошибки БД не должны утекать клиенту
	assert.NotContains(t, w.Body.String(), "db down")
}