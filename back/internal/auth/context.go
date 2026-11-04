package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SessionFromCtx(c *gin.Context) (AuthenticatedSession, bool) {
	v, ok := c.Get(sessionContextKey)
	if !ok {
		return AuthenticatedSession{}, false
	}
	s, ok := v.(AuthenticatedSession)
	if !ok {
		return AuthenticatedSession{}, false
	}
	return s, true
}

func UserIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	s, ok := SessionFromCtx(c)
	if !ok || s.User.ID == uuid.Nil {
		return uuid.Nil, false
	}
	return s.User.ID, true
}