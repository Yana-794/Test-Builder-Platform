package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)

type User struct {
	ID              uuid.UUID
	Email           string
	DisplayName     string
	PasswordHash    string `json:"-"`
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
 type StoredSession struct {
	User      User
	CSRFToken []byte
	ExpiresAt time.Time
}
type AuthenticatedSession struct {
	User      User
	CSRFToken string
	ExpiresAt time.Time
}

type LoginResult struct {
	User         User
	SessionToken string
	CSRFToken    string
	ExpiresAt    time.Time
}