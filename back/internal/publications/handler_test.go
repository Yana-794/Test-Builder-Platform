package publications_test

import (
	"context"
	"encoding/json"
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
	validateFn func(ctx context.Context, ownerID, testID uuid.UUID) (publications.ValidateResponse, error)
	publishFn  func(ctx context.Context, ownerID, testID uuid.UUID, expectedRevision int) (publications.PublishResponse, error)
}

func (m *mockService) Validate(ctx context.Context, ownerID, testID uuid.UUID) (publications.ValidateResponse, error) {
	return m.validateFn(ctx, ownerID, testID)
}
func (m *mockService) Publish(ctx context.Context, ownerID, testID uuid.UUID, expectedRevision int) (publications.PublishResponse, error) {
	return m.publishFn(ctx, ownerID, testID, expectedRevision)
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

func TestHandler_Publish_ValidationFailed(t *testing.T) {
	mock := &mockService{
		publishFn: func(_ context.Context, _, _ uuid.UUID, _ int) (publications.PublishResponse, error) {
			return publications.PublishResponse{}, &publications.ValidationError{
				Details: []publications.FieldDetail{
					{
						Field:   "questions[0].prompt",
						Code:    "REQUIRED",
						Message: "Поле 'Вопрос' обязательно для заполнения",
					},
					{
						Field:   "questions[1].points",
						Code:    "INVALID_POINTS",
						Message: "Баллы за вопрос должны быть больше нуля",
					},
				},
			}
		},
	}

	r := gin.New()
	r.POST("/tests/:testId/publish", withSession(publications.NewHandler(mock).Publish))

	body := `{"expectedRevision":1}`
	req := httptest.NewRequest(http.MethodPost,
		"/tests/"+uuid.NewString()+"/publish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

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
}