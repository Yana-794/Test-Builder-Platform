package tests_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"test-builder/internal/auth"
	"test-builder/internal/tests"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	createFn func(ctx context.Context, ownerID uuid.UUID, req tests.CreateTestRequest) (tests.Test, error)
	getFn    func(ctx context.Context, ownerID, testID uuid.UUID) (tests.Test, []tests.Question, error)
	saveFn   func(ctx context.Context, ownerID, testID uuid.UUID, req tests.SaveDraftRequest) (tests.Test, []tests.Question, error)
}

func (m *mockService) CreateTest(ctx context.Context, ownerID uuid.UUID, req tests.CreateTestRequest) (tests.Test, error) {
	return m.createFn(ctx, ownerID, req)
}
func (m *mockService) GetTestForEditor(ctx context.Context, ownerID, testID uuid.UUID) (tests.Test, []tests.Question, error) {
	return m.getFn(ctx, ownerID, testID)
}
func (m *mockService) SaveDraft(ctx context.Context, ownerID, testID uuid.UUID, req tests.SaveDraftRequest) (tests.Test, []tests.Question, error) {
	return m.saveFn(ctx, ownerID, testID, req)
}

func init() { gin.SetMode(gin.TestMode) }

// withSession кладёт сессию в контекст, имитируя RequireAuth
func withSession(h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("authenticatedSession", auth.AuthenticatedSession{
			User: auth.User{ID: uuid.New()},
		})
		h(c)
	}
}

func sampleTest() tests.Test {
	id := uuid.New()
	owner := uuid.New()
	return tests.Test{
		ID: id, OwnerID: owner,
		Title: "T", Description: "D", Subject: nil,
		Status: "draft", Revision: 2,
	}
}

// POST /tests

func TestHandler_CreateTest_Success(t *testing.T) {
	mock := &mockService{
		createFn: func(_ context.Context, _ uuid.UUID, _ tests.CreateTestRequest) (tests.Test, error) {
			return sampleTest(), nil
		},
	}

	r := gin.New()
	r.POST("/tests", withSession(tests.NewHandler(mock).CreateTest))

	req := httptest.NewRequest(http.MethodPost, "/tests",
		strings.NewReader(`{"title":"T","description":"D","subject":null}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	tObj := body["test"].(map[string]any)
	assert.Equal(t, "T", tObj["title"])
	assert.Equal(t, "draft", tObj["status"])
	assert.Equal(t, float64(2), tObj["revision"])
}

func TestHandler_CreateTest_Unauthorized(t *testing.T) {
	r := gin.New()
	r.POST("/tests", tests.NewHandler(&mockService{}).CreateTest)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/tests",
		strings.NewReader(`{"title":"T","description":"","subject":null}`)))

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "UNAUTHORIZED", body["error"].(map[string]any)["code"])
}

func TestHandler_CreateTest_ValidationError(t *testing.T) {
	mock := &mockService{
		createFn: func(_ context.Context, _ uuid.UUID, _ tests.CreateTestRequest) (tests.Test, error) {
			return tests.Test{}, &tests.ValidationError{
				Details: []tests.FieldDetail{{Field: "title", Message: "is required"}},
			}
		},
	}
	r := gin.New()
	r.POST("/tests", withSession(tests.NewHandler(mock).CreateTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/tests",
		strings.NewReader(`{"title":"","description":"","subject":null}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
	details := errObj["details"].([]any)
	require.Len(t, details, 1)
	assert.Equal(t, "title", details[0].(map[string]any)["field"])
}

// GET /tests/:testId

func TestHandler_GetTest_NotFound(t *testing.T) {
	mock := &mockService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (tests.Test, []tests.Question, error) {
			return tests.Test{}, nil, tests.ErrTestNotFound
		},
	}
	r := gin.New()
	r.GET("/tests/:testId", withSession(tests.NewHandler(mock).GetTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests/"+uuid.NewString(), nil))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["error"].(map[string]any)["code"])
}

func TestHandler_GetTest_InvalidUUID(t *testing.T) {
	mock := &mockService{} 
	r := gin.New()
	r.GET("/tests/:testId", withSession(tests.NewHandler(mock).GetTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests/not-a-uuid", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// PUT /tests/:testId/draft

func TestHandler_SaveDraft_RevisionConflict(t *testing.T) {
	mock := &mockService{
		saveFn: func(_ context.Context, _, _ uuid.UUID, _ tests.SaveDraftRequest) (tests.Test, []tests.Question, error) {
			return tests.Test{}, nil, tests.ErrRevisionConflict
		},
	}
	r := gin.New()
	r.PUT("/tests/:testId/draft", withSession(tests.NewHandler(mock).SaveDraft))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/tests/"+uuid.NewString()+"/draft",
		strings.NewReader(`{"expectedRevision":1,"test":{"title":"T","description":"","subject":null},"questions":[]}`)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "CONFLICT", errObj["code"])
	assert.Contains(t, strings.ToLower(errObj["message"].(string)), "another tab")
}

func TestHandler_SaveDraft_NotDraft(t *testing.T) {
	mock := &mockService{
		saveFn: func(_ context.Context, _, _ uuid.UUID, _ tests.SaveDraftRequest) (tests.Test, []tests.Question, error) {
			return tests.Test{}, nil, tests.ErrTestNotDraft
		},
	}
	r := gin.New()
	r.PUT("/tests/:testId/draft", withSession(tests.NewHandler(mock).SaveDraft))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/tests/"+uuid.NewString()+"/draft",
		strings.NewReader(`{"expectedRevision":1,"test":{"title":"T","description":"","subject":null},"questions":[]}`)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "CONFLICT", errObj["code"])
	assert.Contains(t, strings.ToLower(errObj["message"].(string)), "not a draft")
}

func TestHandler_SaveDraft_InvalidReference(t *testing.T) {
	mock := &mockService{
		saveFn: func(_ context.Context, _, _ uuid.UUID, _ tests.SaveDraftRequest) (tests.Test, []tests.Question, error) {
			return tests.Test{}, nil, tests.ErrInvalidReference
		},
	}
	r := gin.New()
	r.PUT("/tests/:testId/draft", withSession(tests.NewHandler(mock).SaveDraft))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/tests/"+uuid.NewString()+"/draft",
		strings.NewReader(`{"expectedRevision":1,"test":{"title":"T","description":"","subject":null},"questions":[]}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "VALIDATION_ERROR", body["error"].(map[string]any)["code"])
}

func TestHandler_SaveDraft_Unauthorized(t *testing.T) {
	r := gin.New()
	r.PUT("/tests/:testId/draft", tests.NewHandler(&mockService{}).SaveDraft)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/tests/"+uuid.NewString()+"/draft",
		strings.NewReader(`{"expectedRevision":1,"test":{"title":"T","description":"","subject":null},"questions":[]}`)))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

var _ = errors.New