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
	listFn   func(ctx context.Context, ownerID uuid.UUID, filter string) ([]tests.Test, error)
}

func (m *mockService) CreateTest(
	ctx context.Context, ownerID uuid.UUID, req tests.CreateTestRequest,
) (tests.Test, error) {
	return m.createFn(ctx, ownerID, req)
}

func (m *mockService) GetTestForEditor(
	ctx context.Context, ownerID, testID uuid.UUID,
) (tests.Test, []tests.Question, error) {
	return m.getFn(ctx, ownerID, testID)
}

func (m *mockService) SaveDraft(
	ctx context.Context, ownerID, testID uuid.UUID, req tests.SaveDraftRequest,
) (tests.Test, []tests.Question, error) {
	return m.saveFn(ctx, ownerID, testID, req)
}

func (m *mockService) ListTests(
	ctx context.Context, ownerID uuid.UUID, filter string,
) ([]tests.Test, error) {
	return m.listFn(ctx, ownerID, filter)
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

func sampleTest() tests.Test {
	return tests.Test{
		ID:          uuid.New(),
		OwnerID:     uuid.New(),
		Title:       "T",
		Description: "D",
		Subject:     nil,
		Status:      "draft",
		Revision:    2,
	}
}

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

func TestHandler_CreateTest_BindError(t *testing.T) {
	mock := &mockService{
		createFn: func(_ context.Context, _ uuid.UUID, _ tests.CreateTestRequest) (tests.Test, error) {
			t.Fatal("createFn must not be called when bind fails")
			return tests.Test{}, nil
		},
	}
	r := gin.New()
	r.POST("/tests", withSession(tests.NewHandler(mock).CreateTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/tests",
		strings.NewReader(`{"title":`))) // обрезанный JSON

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "VALIDATION_ERROR", body["error"].(map[string]any)["code"])
}

func TestHandler_CreateTest_InternalError(t *testing.T) {
	mock := &mockService{
		createFn: func(_ context.Context, _ uuid.UUID, _ tests.CreateTestRequest) (tests.Test, error) {
			return tests.Test{}, errors.New("db connection lost")
		},
	}
	r := gin.New()
	r.POST("/tests", withSession(tests.NewHandler(mock).CreateTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/tests",
		strings.NewReader(`{"title":"T","description":"","subject":null}`)))

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "INTERNAL_ERROR", body["error"].(map[string]any)["code"])
	assert.NotContains(t, w.Body.String(), "db connection lost")
}



func TestHandler_GetTest_Success(t *testing.T) {
	mock := &mockService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (tests.Test, []tests.Question, error) {
			return sampleTest(), []tests.Question{}, nil
		},
	}
	r := gin.New()
	r.GET("/tests/:testId", withSession(tests.NewHandler(mock).GetTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests/"+uuid.NewString(), nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	tObj := body["test"].(map[string]any)
	assert.Equal(t, "T", tObj["title"])
	assert.Equal(t, "draft", tObj["status"])
	qs, ok := tObj["questions"].([]any)
	require.True(t, ok, "questions must be [] (array), not null")
	assert.Empty(t, qs)
}

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
	mock := &mockService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (tests.Test, []tests.Question, error) {
			t.Fatal("getFn must not be called for invalid UUID")
			return tests.Test{}, nil, nil
		},
	}
	r := gin.New()
	r.GET("/tests/:testId", withSession(tests.NewHandler(mock).GetTest))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests/not-a-uuid", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_SaveDraft_Success(t *testing.T) {
	saved := sampleTest()
	saved.Revision = 3

	mock := &mockService{
		saveFn: func(_ context.Context, _, _ uuid.UUID, _ tests.SaveDraftRequest) (tests.Test, []tests.Question, error) {
			return saved, []tests.Question{}, nil
		},
	}
	r := gin.New()
	r.PUT("/tests/:testId/draft", withSession(tests.NewHandler(mock).SaveDraft))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/tests/"+uuid.NewString()+"/draft",
		strings.NewReader(`{"expectedRevision":2,"test":{"title":"T","description":"","subject":null},"questions":[]}`)))

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	tObj := body["test"].(map[string]any)
	assert.Equal(t, float64(3), tObj["revision"])
}

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


func TestHandler_ListTests_Success(t *testing.T) {
	mock := &mockService{
		listFn: func(_ context.Context, _ uuid.UUID, _ string) ([]tests.Test, error) {
			return []tests.Test{sampleTest()}, nil
		},
	}
	r := gin.New()
	r.GET("/tests", withSession(tests.NewHandler(mock).ListTests))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests?filter=draft", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "draft", body["filter"])
	assert.Equal(t, float64(1), body["total"])

	items, ok := body["tests"].([]any)
	require.True(t, ok, "tests must be array, not null")
	require.Len(t, items, 1)
	first := items[0].(map[string]any)
	assert.Equal(t, "T", first["title"])
	assert.Equal(t, "draft", first["status"])
}

func TestHandler_ListTests_InvalidFilter(t *testing.T) {
	mock := &mockService{
		listFn: func(_ context.Context, _ uuid.UUID, _ string) ([]tests.Test, error) {
			return nil, tests.ErrInvalidFilter
		},
	}
	r := gin.New()
	r.GET("/tests", withSession(tests.NewHandler(mock).ListTests))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tests?filter=xyz", nil))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "VALIDATION_ERROR", body["error"].(map[string]any)["code"])
}