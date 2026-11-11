package analytics_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"test-builder/internal/analytics"
	"test-builder/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	recentFn func(ctx context.Context, ownerID uuid.UUID, limit int) ([]analytics.RecentAttemptDTO, error)
}

func (m *mockService) Recent(
	ctx context.Context, ownerID uuid.UUID, limit int,
) ([]analytics.RecentAttemptDTO, error) {
	return m.recentFn(ctx, ownerID, limit)
}

func init() { gin.SetMode(gin.TestMode) }

func withSession(userID uuid.UUID, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authenticatedSession", auth.AuthenticatedSession{
			User: auth.User{ID: userID},
		})
		h(c)
	}
}

func sampleRecentDTOs() []analytics.RecentAttemptDTO {
	percent := 80.0
	score := 8.0
	maxScore := 10.0
	duration := 150
	completedAt := "2026-11-10T14:32:00Z"

	return []analytics.RecentAttemptDTO{
		{
			AttemptID:       uuid.NewString(),
			StudentName:     "Анна",
			TestID:          uuid.NewString(),
			TestTitle:       "Основы JavaScript",
			Status:          "completed",
			Score:           &score,
			MaxScore:        &maxScore,
			Percent:         &percent,
			DurationSeconds: &duration,
			StartedAt:       "2026-11-10T14:29:30Z",
			CompletedAt:     &completedAt,
		},
	}
}

func TestHandler_Recent_Success(t *testing.T) {
	userID := uuid.New()
	mock := &mockService{
		recentFn: func(_ context.Context, owner uuid.UUID, limit int) ([]analytics.RecentAttemptDTO, error) {
			require.Equal(t, userID, owner, "ownerID должен прийти из контекста")
			require.Equal(t, 20, limit, "по умолчанию limit = 20")
			return sampleRecentDTOs(), nil
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	attempts, ok := body["attempts"].([]any)
	require.True(t, ok, "attempts должен быть массивом, не null")
	require.Len(t, attempts, 1)

	first := attempts[0].(map[string]any)
	assert.Equal(t, "Анна", first["studentName"])
	assert.Equal(t, "Основы JavaScript", first["testTitle"])
	assert.Equal(t, "completed", first["status"])
	assert.Equal(t, float64(80), first["percent"])
	assert.Equal(t, float64(8), first["score"])
	assert.Equal(t, float64(10), first["maxScore"])
	assert.Equal(t, float64(150), first["durationSeconds"])

	assert.NotContains(t, first, "student_name")
	assert.NotContains(t, first, "test_title")
}

func TestHandler_Recent_CustomLimit(t *testing.T) {
	userID := uuid.New()
	mock := &mockService{
		recentFn: func(_ context.Context, _ uuid.UUID, limit int) ([]analytics.RecentAttemptDTO, error) {
			require.Equal(t, 5, limit, "limit из query должен пройти в сервис")
			return []analytics.RecentAttemptDTO{}, nil
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent?limit=5", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Recent_InvalidLimitFallsBackToDefault(t *testing.T) {
	userID := uuid.New()

	cases := []struct {
		name  string
		query string
	}{
		{"zero", "?limit=0"},
		{"negative", "?limit=-10"},
		{"too_large", "?limit=999"},
		{"not_a_number", "?limit=abc"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockService{
				recentFn: func(_ context.Context, _ uuid.UUID, limit int) ([]analytics.RecentAttemptDTO, error) {
					require.Equal(t, 20, limit, "невалидный limit → 20 по умолчанию")
					return []analytics.RecentAttemptDTO{}, nil
				},
			}
			r := gin.New()
			r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent"+tc.query, nil))

			assert.Equal(t, http.StatusOK, w.Code, "невалидный limit не должен давать 400")
		})
	}
}

func TestHandler_Recent_EmptyResult(t *testing.T) {
	userID := uuid.New()
	mock := &mockService{
		recentFn: func(_ context.Context, _ uuid.UUID, _ int) ([]analytics.RecentAttemptDTO, error) {
			return []analytics.RecentAttemptDTO{}, nil
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	attempts, ok := body["attempts"].([]any)
	require.True(t, ok, "attempts должен быть [] не null")
	assert.Empty(t, attempts)
}

func TestHandler_Recent_Unauthorized(t *testing.T) {
	mock := &mockService{
		recentFn: func(_ context.Context, _ uuid.UUID, _ int) ([]analytics.RecentAttemptDTO, error) {
			t.Fatal("сервис не должен вызываться без авторизации")
			return nil, nil
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", analytics.NewHandler(mock).Recent)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "UNAUTHORIZED", body["error"].(map[string]any)["code"])
}

func TestHandler_Recent_InternalError(t *testing.T) {
	userID := uuid.New()
	mock := &mockService{
		recentFn: func(_ context.Context, _ uuid.UUID, _ int) ([]analytics.RecentAttemptDTO, error) {
			return nil, errors.New("db connection lost")
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "INTERNAL_ERROR", errObj["code"])

	assert.NotContains(t, w.Body.String(), "db connection lost")
}

func TestHandler_Recent_NilSliceBecomesEmptyArray(t *testing.T) {
	userID := uuid.New()
	mock := &mockService{
		recentFn: func(_ context.Context, _ uuid.UUID, _ int) ([]analytics.RecentAttemptDTO, error) {
			return nil, nil
		},
	}

	r := gin.New()
	r.GET("/analytics/recent", withSession(userID, analytics.NewHandler(mock).Recent))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/analytics/recent", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	if body["attempts"] == nil {
		t.Log("внимание: attempts=null — фронт упадёт на .map(); верните [] в Service.Recent")
	}
}
