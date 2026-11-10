package publications_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"test-builder/internal/auth"
	"test-builder/internal/publications"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



type mockService struct {
	validateFn     func(ctx context.Context, ownerID, testID uuid.UUID) (publications.ValidateResponse, error)
	publishFn      func(ctx context.Context, ownerID, testID uuid.UUID, expectedRevision int) (publications.PublishResponse, error)
	getPublicationFn func(ctx context.Context, ownerID, testID uuid.UUID) (publications.PublishResponseData, error)
}

func (m *mockService) Validate(ctx context.Context, ownerID, testID uuid.UUID) (publications.ValidateResponse, error) {
	return m.validateFn(ctx, ownerID, testID)
}

func (m *mockService) Publish(ctx context.Context, ownerID, testID uuid.UUID, expectedRevision int) (publications.PublishResponse, error) {
	return m.publishFn(ctx, ownerID, testID, expectedRevision)
}

func (m *mockService) GetPublication(ctx context.Context, ownerID, testID uuid.UUID) (publications.PublishResponseData, error) {
	if m.getPublicationFn == nil {
		return publications.PublishResponseData{}, nil
	}
	return m.getPublicationFn(ctx, ownerID, testID)
}


func init() { gin.SetMode(gin.TestMode) }

func withSession(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authenticatedSession", auth.AuthenticatedSession{
			User: auth.User{ID: uuid.New()},
		})
		h(c)
	}
}

func samplePublishResponse() publications.PublishResponse {
	return publications.PublishResponse{
		Publication: publications.PublishResponseData{
			ID:          uuid.NewString(),
			Version:     1,
			PublicToken: "test-token-abc",
			PublicURL:   "http://localhost/test/test-token-abc",
			PublishedAt: "2026-11-10T14:00:00Z",
		},
	}
}


func TestHandler_Publish_ValidationFailed(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			return publications.PublishResponse{}, &publications.ValidationError{
				Details: []publications.FieldDetail{
					{Field: "questions[0].prompt", Code: "REQUIRED", Message: "Поле 'Вопрос' обязательно"},
					{Field: "questions[1].points", Code: "INVALID_POINTS", Message: "Баллы должны быть больше нуля"},
				},
			}
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	req := httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "PUBLISH_VALIDATION_FAILED", errObj["code"])
	assert.Equal(t, "Test is not ready for publishing", errObj["message"])

	details := errObj["details"].([]any)
	require.Len(t, details, 2, "handler должен вернуть все details из ValidationError")
	assert.Equal(t, "questions[0].prompt", details[0].(map[string]any)["field"])
	assert.Equal(t, "REQUIRED", details[0].(map[string]any)["code"])
	assert.Equal(t, "questions[1].points", details[1].(map[string]any)["field"])
	assert.Equal(t, "INVALID_POINTS", details[1].(map[string]any)["code"])
}



func TestHandler_Publish_Success(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, expectedRevision int) (publications.PublishResponse, error) {
			require.Equal(t, 3, expectedRevision, "expectedRevision должен пройти из запроса в сервис")
			return samplePublishResponse(), nil
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	req := httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	pub := resp["publication"].(map[string]any)
	assert.Equal(t, "test-token-abc", pub["publicToken"])
	assert.Equal(t, "http://localhost/test/test-token-abc", pub["publicUrl"])
	assert.Equal(t, float64(1), pub["version"])
}


func TestHandler_Publish_RevisionConflict(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			return publications.PublishResponse{}, publications.ErrRevisionConflict
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":1}`)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "CONFLICT", errObj["code"])
	assert.Contains(t, strings.ToLower(errObj["message"].(string)), "another tab")
}

func TestHandler_Publish_NotDraft(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			return publications.PublishResponse{}, publications.ErrTestNotDraft
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":1}`)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	errObj := resp["error"].(map[string]any)
	assert.Equal(t, "CONFLICT", errObj["code"])
	assert.Contains(t, strings.ToLower(errObj["message"].(string)), "not a draft")
}

func TestHandler_Publish_TestNotFound(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			return publications.PublishResponse{}, publications.ErrTestNotFound
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":1}`)))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp["error"].(map[string]any)["code"])
}


func TestHandler_Publish_Unauthorized(t *testing.T) {
	r := gin.New()
	r.POST("/tests/:testId/publish", publications.NewHandler(&mockService{}).Publish)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":1}`)))

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp["error"].(map[string]any)["code"])
}

func TestHandler_Publish_BindError(t *testing.T) {
	// Битый JSON — сервис не должен вызываться.
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			t.Fatal("publishFn must not be called on bind error")
			return publications.PublishResponse{}, nil
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["error"].(map[string]any)["code"])
}

func TestHandler_Publish_InvalidRevision(t *testing.T) {
	// expectedRevision < 1 — handler отклоняет до вызова сервиса.
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			t.Fatal("publishFn must not be called for invalid revision")
			return publications.PublishResponse{}, nil
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(`{"expectedRevision":0}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["error"].(map[string]any)["code"])
}


func TestHandler_Validate_Valid(t *testing.T) {
	mock := &mockService{
		validateFn: func(_ context.Context, _, _ uuid.UUID) (publications.ValidateResponse, error) {
			return publications.ValidateResponse{
				Valid:   true,
				Errors:  []publications.FieldDetail{},
				Summary: publications.ValidateSummary{QuestionCount: 3, MaxScore: 5},
			}, nil
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/validate", withSession(publications.NewHandler(mock).Validate))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/validate", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["valid"])
	summary := resp["summary"].(map[string]any)
	assert.Equal(t, float64(3), summary["questionCount"])
	assert.Equal(t, float64(5), summary["maxScore"])
}

func TestHandler_Validate_Invalid(t *testing.T) {
	mock := &mockService{
		validateFn: func(_ context.Context, _, _ uuid.UUID) (publications.ValidateResponse, error) {
			return publications.ValidateResponse{
				Valid: false,
				Errors: []publications.FieldDetail{
					{Field: "questions", Code: "REQUIRED", Message: "Минимум один вопрос"},
				},
				Summary: publications.ValidateSummary{QuestionCount: 0, MaxScore: 0},
			}, nil
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/validate", withSession(publications.NewHandler(mock).Validate))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/validate", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["valid"])
	errs := resp["errors"].([]any)
	require.Len(t, errs, 1)
	assert.Equal(t, "questions", errs[0].(map[string]any)["field"])
}

var _ = errors.New