package attempts

import (
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func cookieName(attemptID uuid.UUID) string {
	return "tbp_attempt_" + attemptID.String()
}

// устанавливает HttpOnly cookie с секретом
func setAttemptCookie(
	c *gin.Context,
	attemptID uuid.UUID,
	secret string,
	expiresAt time.Time,
	secure bool,
) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName(attemptID),
		Value:    secret,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	})
}

// читает секрет из cookie попытки
func readAttemptCookie(c *gin.Context, attemptID uuid.UUID) (string, error) {
	return c.Cookie(cookieName(attemptID))
}

func hashSecret(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}
