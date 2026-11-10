package attempts_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"test-builder/internal/attempts"
	"test-builder/internal/publications"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



type mockService struct {
	getPublicTestFn func(ctx context.Context, token string) (publications.PublicTest, error)
	createAttemptFn func(ctx context.Context, token, studentName string) (attempts.Attempt, string, error)
	authenticateFn  func(ctx context.Context, attemptID uuid.UUID, secret string) (attempts.Attempt, error)
	getStateFn      func(ctx context.Context, attempt attempts.Attempt) (attempts.AttemptStateResponse, error)
	saveAnswersFn   func(ctx context.Context, attempt attempts.Attempt, req attempts.SaveAnswersRequest) error
	submitFn        func(ctx context.Context, attempt attempts.Attempt) (attempts.AttemptResultDTO, error)
	getResultFn     func(ctx context.Context, attempt attempts.Attempt) (attempts.AttemptResultDTO, error)
}

func (m *mockService) GetPublicTest(ctx context.Context, token string) (publications.PublicTest, error) {
	return m.getPublicTestFn(ctx, token)
}
func (m *mockService) CreateAttempt(ctx context.Context, token, studentName string) (attempts.Attempt, string, error) {
	return m.createAttemptFn(ctx, token, studentName)
}
func (m *mockService) AuthenticateAttempt(ctx context.Context, attemptID uuid.UUID, secret string) (attempts.Attempt, error) {
	return m.authenticateFn(ctx, attemptID, secret)
}
func (m *mockService) GetState(ctx context.Context, a attempts.Attempt) (attempts.AttemptStateResponse, error) {
	return m.getStateFn(ctx, a)
}
func (m *mockService) SaveAnswers(ctx context.Context, a attempts.Attempt, req attempts.SaveAnswersRequest) error {
	return m.saveAnswersFn(ctx, a, req)
}
func (m *mockService) Submit(ctx context.Context, a attempts.Attempt) (attempts.AttemptResultDTO, error) {
	return m.submitFn(ctx, a)
}
func (m *mockService) GetResult(ctx context.Context, a attempts.Attempt) (attempts.AttemptResultDTO, error) {
	return m.getResultFn(ctx, a)
}

func init() { gin.SetMode(gin.TestMode) }

// withAttempt имитирует middleware RequireAttemptCookie — кладёт попытку в контекст
func withAttempt(a attempts.Attempt, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("attempt", a)
		h(c)
	}
}

func cookieName(attemptID uuid.UUID) string {
	return "tbp_attempt_" + attemptID.String()
}

func sampleAttempt() attempts.Attempt {
	return attempts.Attempt{
		ID:            uuid.New(),
		PublicationID: uuid.New(),
		StudentName:   "Анна",
		Status:        attempts.StatusInProgress,
		StartedAt:     time.Now().UTC(),
		ExpiresAt:     time.Now().UTC().Add(24 * time.Hour),
	}
}

func completedAttempt() attempts.Attempt {
	a := sampleAttempt()
	a.Status = attempts.StatusCompleted
	now := time.Now().UTC()
	a.CompletedAt = &now
	score := 8.0
	maxScore := 10.0
	percent := 80.0
	duration := 150
	a.Score = &score
	a.MaxScore = &maxScore
	a.ScorePercent = &percent
	a.DurationSeconds = &duration
	return a
}

func samplePublicTest() publications.PublicTest {
	return publications.PublicTest{
		Title:         "Основы JavaScript",
		Description:   "Базовый тест",
		QuestionCount: 1,
		Questions: []publications.PublicQuestion{
			{
				ID:       uuid.New(),
				Prompt:   "1+1",
				Type:     "single_choice",
				Position: 0,
				Options: []publications.PublicOption{
					{ID: uuid.New(), Text: "1"},
					{ID: uuid.New(), Text: "2"},
				},
			},
		},
	}
}

func sampleResultDTO(a attempts.Attempt) attempts.AttemptResultDTO {
	return attempts.AttemptResultDTO{
		ID:              a.ID.String(),
		Status:          attempts.StatusCompleted,
		Score:           8,
		MaxScore:        10,
		Percentage:      80,
		DurationSeconds: 150,
		CompletedAt:     a.CompletedAt.UTC().Format(time.RFC3339),
	}
}

func TestHandler_GetPublicTest_Success(t *testing.T) {
	mock := &mockService{
		getPublicTestFn: func(_ context.Context, token string) (publications.PublicTest, error) {
			require.Equal(t, "test-token-abc", token, "token должен пройти из URL")
			return samplePublicTest(), nil
		},
	}
	r := gin.New()
	r.GET("/public/tests/:token", attempts.NewHandler(mock, "development").GetPublicTest)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/tests/test-token-abc", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	test := body["test"].(map[string]any)
	assert.Equal(t, "Основы JavaScript", test["title"])
	assert.Equal(t, float64(1), test["questionCount"])
	qs := test["questions"].([]any)
	require.Len(t, qs, 1)
	q := qs[0].(map[string]any)
	assert.Equal(t, "1+1", q["prompt"])
	opts := q["options"].([]any)
	require.Len(t, opts, 2)
	first := opts[0].(map[string]any)
	assert.NotContains(t, first, "isCorrect", "SEC: isCorrect не должен попадать в публичный ответ")
}

func TestHandler_GetPublicTest_NotFound(t *testing.T) {
	mock := &mockService{
		getPublicTestFn: func(_ context.Context, _ string) (publications.PublicTest, error) {
			return publications.PublicTest{}, attempts.ErrPublicationNotFound
		},
	}
	r := gin.New()
	r.GET("/public/tests/:token", attempts.NewHandler(mock, "development").GetPublicTest)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/tests/missing", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["error"].(map[string]any)["code"])
}

func TestHandler_CreateAttempt_Success(t *testing.T) {
	a := sampleAttempt()
	mock := &mockService{
		createAttemptFn: func(_ context.Context, token, name string) (attempts.Attempt, string, error) {
			require.Equal(t, "test-token-abc", token)
			require.Equal(t, "Анна", name)
			return a, "secret-abc", nil
		},
	}
	r := gin.New()
	r.POST("/public/tests/:token/attempts", attempts.NewHandler(mock, "development").CreateAttempt)

	req := httptest.NewRequest(http.MethodPost, "/public/tests/test-token-abc/attempts",
		strings.NewReader(`{"studentName":"Анна"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 201 + Set-Cookie
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1, "должна ставиться одна cookie")
	assert.Equal(t, cookieName(a.ID), cookies[0].Name)
	assert.Equal(t, "secret-abc", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly, "SEC: HttpOnly обязателен")
	assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	att := body["attempt"].(map[string]any)
	assert.Equal(t, a.ID.String(), att["id"])
	assert.Equal(t, "in_progress", att["status"])
}

func TestHandler_CreateAttempt_ValidationError(t *testing.T) {
	mock := &mockService{
		createAttemptFn: func(_ context.Context, _, _ string) (attempts.Attempt, string, error) {
			return attempts.Attempt{}, "", &attempts.ValidationError{
				Details: []attempts.FieldDetail{
					{Field: "studentName", Code: "REQUIRED", Message: "Имя обязательно"},
				},
			}
		},
	}
	r := gin.New()
	r.POST("/public/tests/:token/attempts", attempts.NewHandler(mock, "development").CreateAttempt)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/public/tests/t/attempts",
		strings.NewReader(`{"studentName":""}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
	details := errObj["details"].([]any)
	require.Len(t, details, 1)
	assert.Equal(t, "studentName", details[0].(map[string]any)["field"])

	assert.Empty(t, w.Result().Cookies())
}

func TestHandler_CreateAttempt_PublicationNotFound(t *testing.T) {
	mock := &mockService{
		createAttemptFn: func(_ context.Context, _, _ string) (attempts.Attempt, string, error) {
			return attempts.Attempt{}, "", attempts.ErrPublicationNotFound
		},
	}
	r := gin.New()
	r.POST("/public/tests/:token/attempts", attempts.NewHandler(mock, "development").CreateAttempt)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/public/tests/missing/attempts",
		strings.NewReader(`{"studentName":"Анна"}`)))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_GetState_Success(t *testing.T) {
	a := sampleAttempt()
	mock := &mockService{
		getStateFn: func(_ context.Context, passed attempts.Attempt) (attempts.AttemptStateResponse, error) {
			require.Equal(t, a.ID, passed.ID, "attempt должен прийти из контекста")
			return attempts.AttemptStateResponse{
				Attempt: attempts.AttemptShortDTO{
					ID: a.ID.String(), Status: a.Status,
					StartedAt: a.StartedAt.UTC().Format(time.RFC3339),
				},
				Test:    samplePublicTest(),
				Answers: []attempts.AnswerDTO{},
			}, nil
		},
	}
	r := gin.New()
	r.GET("/public/attempts/:attemptId", withAttempt(a, attempts.NewHandler(mock, "development").GetState))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/attempts/"+a.ID.String(), nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	test := body["test"].(map[string]any)
	assert.Equal(t, "Основы JavaScript", test["title"])
	assert.Equal(t, float64(1), test["questionCount"])

	answers, ok := body["answers"].([]any)
	require.True(t, ok, "answers должен быть [] (array), не null")
	assert.Empty(t, answers)
}

func TestHandler_SaveAnswers_Success(t *testing.T) {
	a := sampleAttempt()
	mock := &mockService{
		saveAnswersFn: func(_ context.Context, passed attempts.Attempt, req attempts.SaveAnswersRequest) error {
			require.Equal(t, a.ID, passed.ID)
			require.Len(t, req.Answers, 1)
			return nil
		},
	}
	r := gin.New()
	r.PUT("/public/attempts/:attemptId/answers",
		withAttempt(a, attempts.NewHandler(mock, "development").SaveAnswers))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/public/attempts/"+a.ID.String()+"/answers",
		strings.NewReader(`{"answers":[{"questionId":"`+uuid.NewString()+`","answerText":"yes"}]}`)))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestHandler_SaveAnswers_AttemptCompleted(t *testing.T) {
	mock := &mockService{
		saveAnswersFn: func(_ context.Context, _ attempts.Attempt, _ attempts.SaveAnswersRequest) error {
			return attempts.ErrAttemptCompleted
		},
	}
	r := gin.New()
	r.PUT("/public/attempts/:attemptId/answers",
		withAttempt(sampleAttempt(), attempts.NewHandler(mock, "development").SaveAnswers))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/public/attempts/"+uuid.NewString()+"/answers",
		strings.NewReader(`{"answers":[]}`)))

	assert.Equal(t, http.StatusConflict, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "CONFLICT", body["error"].(map[string]any)["code"])
}

func TestHandler_SaveAnswers_Expired(t *testing.T) {
	mock := &mockService{
		saveAnswersFn: func(_ context.Context, _ attempts.Attempt, _ attempts.SaveAnswersRequest) error {
			return attempts.ErrAttemptExpired
		},
	}
	r := gin.New()
	r.PUT("/public/attempts/:attemptId/answers",
		withAttempt(sampleAttempt(), attempts.NewHandler(mock, "development").SaveAnswers))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPut,
		"/public/attempts/"+uuid.NewString()+"/answers",
		strings.NewReader(`{"answers":[]}`)))

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_Submit_Success(t *testing.T) {
	a := completedAttempt()
	mock := &mockService{
		submitFn: func(_ context.Context, passed attempts.Attempt) (attempts.AttemptResultDTO, error) {
			require.Equal(t, a.ID, passed.ID)
			return sampleResultDTO(a), nil
		},
	}
	r := gin.New()
	r.POST("/public/attempts/:attemptId/submit",
		withAttempt(a, attempts.NewHandler(mock, "development").Submit))   // ← было sampleAttempt()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/public/attempts/"+a.ID.String()+"/submit", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	att := body["attempt"].(map[string]any)
	assert.Equal(t, "completed", att["status"])
	assert.Equal(t, float64(8), att["score"])
	assert.Equal(t, float64(10), att["maxScore"])
	assert.Equal(t, float64(80), att["percentage"])
	assert.Equal(t, float64(150), att["durationSeconds"])
}

func TestHandler_Submit_Expired(t *testing.T) {
	mock := &mockService{
		submitFn: func(_ context.Context, _ attempts.Attempt) (attempts.AttemptResultDTO, error) {
			return attempts.AttemptResultDTO{}, attempts.ErrAttemptExpired
		},
	}
	r := gin.New()
	r.POST("/public/attempts/:attemptId/submit",
		withAttempt(sampleAttempt(), attempts.NewHandler(mock, "development").Submit))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost,
		"/public/attempts/"+uuid.NewString()+"/submit", nil))

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_GetResult_Success(t *testing.T) {
	a := completedAttempt()
	mock := &mockService{
		getResultFn: func(_ context.Context, passed attempts.Attempt) (attempts.AttemptResultDTO, error) {
			require.Equal(t, a.ID, passed.ID)
			return sampleResultDTO(a), nil
		},
	}
	r := gin.New()
	r.GET("/public/attempts/:attemptId/result",
		withAttempt(a, attempts.NewHandler(mock, "development").GetResult))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/public/attempts/"+a.ID.String()+"/result", nil))

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	att := body["attempt"].(map[string]any)
	assert.Equal(t, "completed", att["status"])
	assert.Equal(t, float64(80), att["percentage"])
}

func TestHandler_GetResult_NotCompleted(t *testing.T) {
	mock := &mockService{
		getResultFn: func(_ context.Context, _ attempts.Attempt) (attempts.AttemptResultDTO, error) {
			return attempts.AttemptResultDTO{}, attempts.ErrAttemptCompleted
		},
	}
	r := gin.New()
	r.GET("/public/attempts/:attemptId/result",
		withAttempt(sampleAttempt(), attempts.NewHandler(mock, "development").GetResult))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
		"/public/attempts/"+uuid.NewString()+"/result", nil))

	assert.Equal(t, http.StatusConflict, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "CONFLICT", body["error"].(map[string]any)["code"])
}

func TestRequireAttemptCookie_NoCookie(t *testing.T) {
	mock := &mockService{
		authenticateFn: func(_ context.Context, _ uuid.UUID, _ string) (attempts.Attempt, error) {
			t.Fatal("AuthenticateAttempt не должен вызываться без cookie")
			return attempts.Attempt{}, nil
		},
	}
	h := attempts.NewHandler(mock, "development")

	r := gin.New()
	r.GET("/public/attempts/:attemptId",
		h.RequireAttemptCookie(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/public/attempts/"+uuid.NewString(), nil))

	assert.Equal(t, http.StatusNotFound, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "NOT_FOUND", body["error"].(map[string]any)["code"])
}

func TestRequireAttemptCookie_InvalidAttemptID(t *testing.T) {
	mock := &mockService{}
	h := attempts.NewHandler(mock, "development")

	r := gin.New()
	r.GET("/public/attempts/:attemptId",
		h.RequireAttemptCookie(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	req := httptest.NewRequest(http.MethodGet, "/public/attempts/not-a-uuid", nil)
	req.AddCookie(&http.Cookie{Name: cookieName(uuid.New()), Value: "x"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRequireAttemptCookie_AuthFailed(t *testing.T) {
	mock := &mockService{
		authenticateFn: func(_ context.Context, _ uuid.UUID, _ string) (attempts.Attempt, error) {
			return attempts.Attempt{}, attempts.ErrAttemptNotFound
		},
	}
	h := attempts.NewHandler(mock, "development")

	r := gin.New()
	r.GET("/public/attempts/:attemptId",
		h.RequireAttemptCookie(),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)

	attemptID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/public/attempts/"+attemptID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName(attemptID), Value: "wrong-secret"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code,
		"SEC: при неверном секрете — 404, а не 401 (не раскрываем существование)")
}

func TestRequireAttemptCookie_Success(t *testing.T) {
	a := sampleAttempt()
	mock := &mockService{
		authenticateFn: func(_ context.Context, attemptID uuid.UUID, secret string) (attempts.Attempt, error) {
			require.Equal(t, a.ID, attemptID)
			require.Equal(t, "valid-secret", secret)
			return a, nil
		},
	}
	h := attempts.NewHandler(mock, "development")

	var gotFromCtx attempts.Attempt
	r := gin.New()
	r.GET("/public/attempts/:attemptId",
		h.RequireAttemptCookie(),
		func(c *gin.Context) {
			v, _ := c.Get("attempt")
			gotFromCtx = v.(attempts.Attempt)
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/public/attempts/"+a.ID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName(a.ID), Value: "valid-secret"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, a.ID, gotFromCtx.ID, "attempt должен попасть в контекст под ключом attempt")
}
