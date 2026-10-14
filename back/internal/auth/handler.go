package auth

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionContextKey = "authenticatedSession"

type Handler struct {
	service      *Service
	cookieName   string
	secureCookie bool
}

func NewHandler(service *Service, appEnv string) *Handler {
	cookieName := "tbp_session"
	secure := appEnv == "production"

	if secure {
		cookieName = "__Host-tbp_session"
	}

	return &Handler{
		service:      service,
		cookieName:   cookieName,
		secureCookie: secure,
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// func bindError(c *gin.Context, err error) {
// 	var maxBytesErr *http.MaxBytesError
// 	if errors.As(err, &maxBytesErr) {
// 		writeError(c, http.StatusRequestEntityTooLarge,
// 			"REQUEST_TOO_LARGE", "Request body is too large")
// 		return
// 	}

//		// writeError(c, http.StatusBadRequest,
//		// 	"INVALID_REQUEST", "Invalid JSON or unexpected fields")
//		writeError(c, http.StatusBadRequest,
//			"INVALID_REQUEST", err.Error())
//	}
func bindError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		writeError(c, http.StatusRequestEntityTooLarge,
			"REQUEST_TOO_LARGE", "Request body is too large")
		return
	}

	// ВРЕМЕННО 
	writeError(c, http.StatusBadRequest,
		"INVALID_REQUEST", err.Error())
}

func (h *Handler) Register(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	user, err := h.service.Register(
		c.Request.Context(),
		req.Email,
		req.DisplayName,
		req.Password,
	)

	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(c, http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Check the email, name and password. Password must contain at least 12 characters and no more than 128 bytes.")
		return

	case errors.Is(err, ErrEmailExists):
		writeError(c, http.StatusConflict,
			"EMAIL_ALREADY_REGISTERED",
			"An account with this email may already exist.")
		return

	case err != nil:
		slog.ErrorContext(c.Request.Context(),
			"registration failed", "error", err)
		writeError(c, http.StatusInternalServerError,
			"INTERNAL_ERROR", "Something went wrong")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": toUserResponse(user),
	})
}

func (h *Handler) Login(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	result, err := h.service.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)

	if errors.Is(err, ErrInvalidCredentials) {
		writeError(c, http.StatusUnauthorized,
			"INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	if err != nil {
		slog.ErrorContext(c.Request.Context(),
			"login failed", "error", err)
		writeError(c, http.StatusInternalServerError,
			"INTERNAL_ERROR", "Something went wrong")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    result.SessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  result.ExpiresAt,
		MaxAge:   int(time.Until(result.ExpiresAt).Seconds()),
	})

	c.JSON(http.StatusOK, AuthResponse{
		User:      toUserResponse(result.User),
		CSRFToken: result.CSRFToken,
	})
}

func (h *Handler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")

		token, err := c.Cookie(h.cookieName)
		if err != nil {
			writeError(c, http.StatusUnauthorized,
				"UNAUTHENTICATED", "Please sign in")
			c.Abort()
			return
		}

		session, err := h.service.Authenticate(
			c.Request.Context(), token,
		)
		if errors.Is(err, ErrUnauthorized) {
			writeError(c, http.StatusUnauthorized,
				"UNAUTHENTICATED", "Please sign in")
			c.Abort()
			return
		}
		if err != nil {
			slog.ErrorContext(c.Request.Context(),
				"session validation failed", "error", err)
			writeError(c, http.StatusInternalServerError,
				"INTERNAL_ERROR", "Something went wrong")
			c.Abort()
			return
		}

		// Для всех небезопасных HTTP-методов нужен CSRF-токен
		method := c.Request.Method
		unsafe := method != http.MethodGet &&
			method != http.MethodHead &&
			method != http.MethodOptions

		if unsafe && subtle.ConstantTimeCompare(
			[]byte(c.GetHeader("X-CSRF-Token")),
			[]byte(session.CSRFToken),
		) != 1 {
			writeError(c, http.StatusForbidden,
				"CSRF_VALIDATION_FAILED", "Request verification failed")
			c.Abort()
			return
		}

		c.Set(sessionContextKey, session)
		c.Next()
	}
}

func (h *Handler) Me(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	session := c.MustGet(sessionContextKey).(AuthenticatedSession)

	c.JSON(http.StatusOK, AuthResponse{
		User:      toUserResponse(session.User),
		CSRFToken: session.CSRFToken,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	token, err := c.Cookie(h.cookieName)
	if err != nil {
		writeError(c, http.StatusUnauthorized,
			"UNAUTHENTICATED", "Please sign in")
		return
	}

	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		slog.ErrorContext(c.Request.Context(),
			"logout failed", "error", err)
		writeError(c, http.StatusInternalServerError,
			"INTERNAL_ERROR", "Something went wrong")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
	})

	c.Status(http.StatusNoContent)
}
